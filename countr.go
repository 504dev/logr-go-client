package logr_go_client

import (
	"encoding/json"
	"github.com/504dev/logr-go-client/types"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	psnet "github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	"log"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

var ts = time.Now()

type State map[string]*types.Count

func (cm State) String() string {
	text, _ := json.MarshalIndent(cm, "", "  ")
	return string(text)
}

type Counter struct {
	*Config
	net.Conn
	sync.RWMutex
	*time.Ticker
	State
	statePrev    State
	Logname      string
	watchSystem  bool
	watchProcess bool
}

func (co *Counter) connect() error {
	var err error
	if co.Conn == nil {
		co.Conn, err = net.Dial("udp", co.Config.Udp)
	}
	return err
}

func (co *Counter) run(interval time.Duration) {
	co.Ticker = time.NewTicker(interval)
	go (func() {
		for {
			<-co.Ticker.C
			co.Flush()
		}
	})()
}

func (co *Counter) Flush() State {
	if co.watchSystem {
		co.collectSystemInfo()
	}
	if co.watchProcess {
		co.collectProcessInfo()
	}

	co.Lock()
	defer co.Unlock()

	tmp := co.State
	co.statePrev = tmp
	co.State = make(State)

	go func() {
		for _, c := range tmp {
			_, err := co.writeCount(c)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	return tmp
}

func (co *Counter) writeCount(count *types.Count) (int, error) {
	if co.Conn == nil {
		return 0, nil
	}
	lp := types.LogPackage{
		DashId:    co.Config.DashId,
		PublicKey: co.Config.PublicKey,
		Count:     count,
	}
	if !co.Config.NoCipher {
		err := lp.EncryptCount(co.Config.PrivateKey)
		if err != nil {
			return 0, err
		}
	}

	msg, err := json.Marshal(lp)
	if err != nil {
		return 0, err
	}
	_, err = co.Conn.Write(msg)
	if err != nil {
		return 0, err
	}
	return len(msg), nil
}

func (co *Counter) Touch(key string) *types.Count {
	res, _ := co.touchSafe(key)
	return res
}

func (co *Counter) touchSafe(key string) (c *types.Count, new bool) {
	co.Lock()
	defer co.Unlock()
	if _, ok := co.State[key]; !ok {
		co.State[key] = &types.Count{
			DashId:   co.Config.DashId,
			Hostname: co.GetHostname(),
			Logname:  co.Logname,
			Keyname:  key,
			Version:  co.GetVersion(),
		}
		new = true
	}
	return co.State[key], new
}

func (co *Counter) Inc(key string, num float64) *types.Count {
	return co.Touch(key).Inc(num)
}

func (co *Counter) IncDiff(key string, num float64) *types.Count {
	res := co.Touch(key)
	if res.Metrics.Inc == nil {
		if prev := co.prevInc(key); prev != nil {
			res.IncLast(prev.Last)
		} else {
			return res.IncLast(num)
		}
	}

	return res.Inc(num).IncLast(num)
}

func (co *Counter) prevInc(key string) *types.Inc {
	co.RLock()
	defer co.RUnlock()
	if co.statePrev != nil && co.statePrev[key] != nil && co.statePrev[key].Metrics.Inc != nil {
		return co.statePrev[key].Metrics.Inc
	}
	return nil
}

func (co *Counter) Max(key string, num float64) *types.Count {
	return co.Touch(key).Max(num)
}

func (co *Counter) Min(key string, num float64) *types.Count {
	return co.Touch(key).Min(num)
}

func (co *Counter) Avg(key string, num float64) *types.Count {
	return co.Touch(key).Avg(num)
}

func (co *Counter) prevAvg(key string) *types.Avg {
	co.RLock()
	defer co.RUnlock()
	if co.statePrev != nil && co.statePrev[key] != nil && co.statePrev[key].Metrics.Avg != nil {
		return co.statePrev[key].Metrics.Avg
	}
	return nil
}

func (co *Counter) Per(key string, taken float64, total float64) *types.Count {
	return co.Touch(key).Per(taken, total)
}

func (co *Counter) Time(key string, d time.Duration) func() time.Duration {
	return co.Touch(key).Time(d)
}

func (co *Counter) Duration() func() time.Duration {
	mtx := sync.Mutex{}
	ts := time.Now()
	return func() time.Duration {
		mtx.Lock()
		defer mtx.Unlock()
		delta := time.Since(ts)
		ts = ts.Add(delta)
		return delta
	}
}

func (co *Counter) DurationFloat64(d time.Duration) func() float64 {
	delta := co.Duration()
	return func() float64 {
		return float64(delta()) / float64(d)
	}
}

func (co *Counter) Snippet(kind string, keyname string, limit int) string {
	w := struct {
		Widget   string `json:"widget"`
		Logname  string `json:"logname"`
		Hostname string `json:"hostname"`
		Keyname  string `json:"keyname"`
		Kind     string `json:"kind"`
		Limit    int    `json:"limit,omitempty"`
	}{
		"counter",
		co.Logname,
		co.GetHostname(),
		keyname,
		kind,
		limit,
	}
	text, _ := json.Marshal(w)
	return string(text)
}

func (co *Counter) collectSystemInfo() {
	l, _ := load.Avg()
	m, _ := mem.VirtualMemory()
	d, _ := disk.Usage("/")
	c, _ := cpu.Percent(time.Second, true)
	co.Avg("la", l.Load1)
	co.Per("mem", float64(m.Used), float64(m.Total))
	co.Per("disk", float64(d.Used), float64(d.Total))
	for _, v := range c {
		co.Per("cpu", v, 100)
	}
	if connections, err := psnet.Connections("inet"); err == nil {
		co.Max("net:inet", float64(len(connections)))
	}
	if connections, err := psnet.Connections("tcp"); err == nil {
		co.Max("net:tcp", float64(len(connections)))
	}
	if connections, err := psnet.Connections("udp"); err == nil {
		co.Max("net:udp", float64(len(connections)))
	}
}
func (co *Counter) WatchSystem() {
	co.watchSystem = true
}

func (co *Counter) collectProcessInfo() {
	proc := process.Process{Pid: int32(os.Getpid())}
	var memState runtime.MemStats
	runtime.ReadMemStats(&memState)
	co.Avg("runtime.NumGoroutine()", float64(runtime.NumGoroutine()))
	co.Avg("runtime.ReadMemStats().HeapObjects", float64(memState.HeapObjects))
	co.Avg("runtime.ReadMemStats().HeapAlloc", float64(memState.HeapAlloc))
	co.Avg("runtime.ReadMemStats().NextGC", float64(memState.NextGC))
	co.IncDiff("runtime.ReadMemStats().TotalAlloc", float64(memState.TotalAlloc))
	co.IncDiff("runtime.ReadMemStats().NumGC", float64(memState.NumGC))
	if cpuPercent, err := proc.CPUPercent(); err == nil {
		co.Per("process.CPUPercent()", cpuPercent/float64(runtime.NumCPU()), 100)
	}
	if memoryPercent, err := proc.MemoryPercent(); err == nil {
		co.Per("process.MemoryPercent()", float64(memoryPercent), 100)
	}
	if numThreads, err := proc.NumThreads(); err == nil {
		co.Avg("process.NumThreads()", float64(numThreads))
	}
	if memoryInfo, err := proc.MemoryInfo(); err == nil {
		co.Avg("process.MemoryInfo().rss", float64(memoryInfo.RSS))
		co.Avg("process.MemoryInfo().vms", float64(memoryInfo.VMS))
	}
	pid := int32(os.Getpid())
	if connections, err := psnet.ConnectionsPid("inet", pid); err == nil {
		co.Avg("process.Connections().inet", float64(len(connections)))
	}
	if connections, err := psnet.ConnectionsPid("tcp", pid); err == nil {
		co.Avg("process.Connections().tcp", float64(len(connections)))
	}
	if connections, err := psnet.ConnectionsPid("udp", pid); err == nil {
		co.Avg("process.Connections().udp", float64(len(connections)))
	}
	co.Max("lifetime", time.Now().Sub(ts).Minutes())
}

func (co *Counter) WatchProcess() {
	co.watchProcess = true
}
