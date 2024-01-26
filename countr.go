package logr_go_client

import (
	"encoding/json"
	"github.com/504dev/logr-go-client/types"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"log"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

var ts = time.Now()

type Tmp map[string]*types.Count

func (cm Tmp) String() string {
	text, _ := json.MarshalIndent(cm, "", "  ")
	return string(text)
}

type Counter struct {
	*Config
	net.Conn
	sync.Mutex
	*time.Ticker
	Tmp
	Logname      string
	watchSystem  bool
	watchProcess bool
}

func (cntr *Counter) connect() error {
	var err error
	if cntr.Conn == nil {
		cntr.Conn, err = net.Dial("udp", cntr.Udp)
	}
	return err
}

func (cntr *Counter) run(interval time.Duration) {
	cntr.Ticker = time.NewTicker(interval)
	go (func() {
		for {
			<-cntr.Ticker.C
			cntr.Flush()
		}
	})()
}

func (cntr *Counter) Flush() Tmp {
	if cntr.watchSystem {
		cntr.collectSystemInfo()
	}
	if cntr.watchProcess {
		cntr.collectProcessInfo()
	}
	cntr.Lock()
	tmp := cntr.Tmp
	cntr.Tmp = make(Tmp)
	cntr.Unlock()
	for _, c := range tmp {
		_, err := cntr.writeCount(c)
		if err != nil {
			log.Println(err)
		}
	}
	return tmp
}

func (cntr *Counter) writeCount(count *types.Count) (int, error) {
	if cntr.Conn == nil {
		return 0, nil
	}
	lp := types.LogPackage{
		DashId:    cntr.Config.DashId,
		PublicKey: cntr.Config.PublicKey,
		Count:     count,
	}
	if !cntr.NoCipher {
		err := lp.EncryptCount(cntr.PrivateKey)
		if err != nil {
			return 0, err
		}
	}

	msg, err := json.Marshal(lp)
	if err != nil {
		return 0, err
	}
	_, err = cntr.Conn.Write(msg)
	if err != nil {
		return 0, err
	}
	return len(msg), nil
}

func (cntr *Counter) Touch(key string) *types.Count {
	cntr.Lock()
	defer cntr.Unlock()
	if _, ok := cntr.Tmp[key]; !ok {
		cntr.Tmp[key] = &types.Count{
			DashId:   cntr.Config.DashId,
			Hostname: cntr.GetHostname(),
			Logname:  cntr.Logname,
			Keyname:  key,
			Version:  cntr.GetVersion(),
		}
	}
	return cntr.Tmp[key]
}
func (cntr *Counter) Inc(key string, num float64) *types.Count {
	return cntr.Touch(key).Inc(num)
}

func (cntr *Counter) Max(key string, num float64) *types.Count {
	return cntr.Touch(key).Max(num)
}

func (cntr *Counter) Min(key string, num float64) *types.Count {
	return cntr.Touch(key).Min(num)
}

func (cntr *Counter) Avg(key string, num float64) *types.Count {
	return cntr.Touch(key).Avg(num)
}

func (cntr *Counter) Per(key string, taken float64, total float64) *types.Count {
	return cntr.Touch(key).Per(taken, total)
}

func (cntr *Counter) Time(key string, d time.Duration) func() time.Duration {
	return cntr.Touch(key).Time(d)
}

func (cntr *Counter) Duration() func() time.Duration {
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

func (cntr *Counter) DurationFloat64(d time.Duration) func() float64 {
	delta := cntr.Duration()
	return func() float64 {
		return float64(delta()) / float64(d)
	}
}

func (cntr *Counter) Snippet(kind string, keyname string, limit int) string {
	w := struct {
		Widget   string `json:"widget"`
		Logname  string `json:"logname"`
		Hostname string `json:"hostname"`
		Keyname  string `json:"keyname"`
		Kind     string `json:"kind"`
		Limit    int    `json:"limit,omitempty"`
	}{
		"counter",
		cntr.Logname,
		cntr.GetHostname(),
		keyname,
		kind,
		limit,
	}
	text, _ := json.Marshal(w)
	return string(text)
}

func (cntr *Counter) collectSystemInfo() {
	l, _ := load.Avg()
	m, _ := mem.VirtualMemory()
	d, _ := disk.Usage("/")
	c, _ := cpu.Percent(time.Second, true)
	cntr.Avg("la", l.Load1)
	cntr.Per("mem", float64(m.Used), float64(m.Total))
	cntr.Per("disk", float64(d.Used), float64(d.Total))
	for _, v := range c {
		cntr.Per("cpu", v, 100)
	}
}
func (cntr *Counter) WatchSystem() {
	cntr.watchSystem = true
}

func (cntr *Counter) collectProcessInfo() {
	proc := process.Process{Pid: int32(os.Getpid())}
	var memState runtime.MemStats
	runtime.ReadMemStats(&memState)
	cntr.Avg("runtime.NumGoroutine()", float64(runtime.NumGoroutine()))
	cntr.Avg("runtime.ReadMemStats().HeapAlloc", float64(memState.HeapAlloc))
	cntr.Avg("runtime.ReadMemStats().HeapObjects", float64(memState.HeapObjects))
	if cpuPercent, err := proc.CPUPercent(); err == nil {
		cntr.Avg("process.CPUPercent()", cpuPercent)
	}
	if numThreads, err := proc.NumThreads(); err == nil {
		cntr.Avg("process.NumThreads()", float64(numThreads))
	}
	if memoryPercent, err := proc.MemoryPercent(); err == nil {
		cntr.Avg("process.MemoryPercent()", float64(memoryPercent))
	}
	if memoryInfo, err := proc.MemoryInfo(); err == nil {
		cntr.Avg("process.MemoryInfo().rss", float64(memoryInfo.RSS))
		cntr.Avg("process.MemoryInfo().vms", float64(memoryInfo.VMS))
	}
	cntr.Avg("execTime", time.Now().Sub(ts).Seconds())
}

func (cntr *Counter) WatchProcess() {
	cntr.watchProcess = true
}
