package logr_go_client

import (
	"encoding/json"
	"github.com/504dev/logr/types"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"log"
	"net"
	"sync"
	"time"
)

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
}

func (ctr *Counter) run(interval time.Duration) error {
	var err error
	if ctr.Conn == nil {
		ctr.Conn, err = net.Dial("udp", ctr.Udp)
	}

	ctr.Ticker = time.NewTicker(interval)
	go (func() {
		for {
			<-ctr.Ticker.C

			ctr.Flush()
		}
	})()

	return err
}

func (ctr *Counter) Flush() Tmp {
	if ctr.watchSystem {
		l, _ := load.Avg()
		m, _ := mem.VirtualMemory()
		d, _ := disk.Usage("/")
		c, _ := cpu.Percent(time.Second, true)
		ctr.Avg("la", l.Load1)
		ctr.Per("mem", float64(m.Used), float64(m.Total))
		ctr.Per("disk", float64(d.Used), float64(d.Total))
		for _, v := range c {
			ctr.Per("cpu", v, 100)
		}
	}
	ctr.Lock()
	tmp := ctr.Tmp
	ctr.Tmp = make(Tmp)
	ctr.Unlock()
	for _, c := range tmp {
		_, err := ctr.writeCount(c)
		if err != nil {
			log.Println(err)
		}
	}
	return tmp
}

func (ctr *Counter) writeCount(count *types.Count) (int, error) {
	if ctr.Conn == nil {
		return 0, nil
	}
	cipherText, err := count.Encrypt(ctr.PrivateKey)
	if err != nil {
		return 0, err
	}
	lp := types.LogPackage{
		DashId:      ctr.Config.DashId,
		PublicKey:   ctr.Config.PublicKey,
		CipherCount: cipherText,
	}
	msg, err := json.Marshal(lp)
	if err != nil {
		return 0, err
	}
	_, err = ctr.Conn.Write(msg)
	if err != nil {
		return 0, err
	}
	return len(msg), nil
}

func (ctr *Counter) Touch(key string) *types.Count {
	ctr.Lock()
	if _, ok := ctr.Tmp[key]; !ok {
		ctr.Tmp[key] = &types.Count{
			DashId:   ctr.Config.DashId,
			Hostname: ctr.GetHostname(),
			Logname:  ctr.Logname,
			Keyname:  key,
			Version:  ctr.GetVersion(),
		}
	}
	ctr.Unlock()
	return ctr.Tmp[key]
}
func (ctr *Counter) Inc(key string, num float64) *types.Count {
	return ctr.Touch(key).Inc(num)
}

func (ctr *Counter) Max(key string, num float64) *types.Count {
	return ctr.Touch(key).Max(num)
}

func (ctr *Counter) Min(key string, num float64) *types.Count {
	return ctr.Touch(key).Min(num)
}

func (ctr *Counter) Avg(key string, num float64) *types.Count {
	return ctr.Touch(key).Avg(num)
}

func (ctr *Counter) Per(key string, taken float64, total float64) *types.Count {
	return ctr.Touch(key).Per(taken, total)
}

func (ctr *Counter) Time(key string, d time.Duration) func() time.Duration {
	return ctr.Touch(key).Time(d)
}

func (ctr *Counter) Widget(kind string, keyname string, limit int) string {
	w := struct {
		Widget   string `json:"widget"`
		Logname  string `json:"logname"`
		Hostname string `json:"hostname"`
		Keyname  string `json:"keyname"`
		Kind     string `json:"kind"`
		Limit    int    `json:"limit,omitempty"`
	}{
		"counter",
		ctr.Logname,
		ctr.GetHostname(),
		keyname,
		kind,
		limit,
	}
	text, _ := json.Marshal(w)
	return string(text)
}

func (ctr *Counter) WatchSystem() {
	ctr.watchSystem = true
}
