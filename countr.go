package go_kidlog

import (
	"encoding/json"
	"github.com/504dev/kidlog/types"
	"log"
	"net"
	"sync"
	"time"
)

func (c *Config) NewCounter(name string) (*Counter, error) {
	conn, err := net.Dial("udp", c.Udp)
	ctr := &Counter{
		Config:  c,
		Conn:    conn,
		Logname: name,
		Tmp:     make(map[string]*types.Count),
	}
	ctr.run(10 * time.Second)
	return ctr, err
}

type Counter struct {
	*Config
	net.Conn
	sync.Mutex
	*time.Ticker
	Logname string
	Tmp     map[string]*types.Count
}

func (ctr *Counter) run(interval time.Duration) {
	ctr.Ticker = time.NewTicker(interval)
	go (func() {
		for {
			<-ctr.Ticker.C
			ctr.flush()
		}
	})()
}

func (ctr *Counter) flush() {
	ctr.Lock()
	tmp := ctr.Tmp
	ctr.Tmp = make(map[string]*types.Count)
	ctr.Unlock()
	for _, c := range tmp {
		_, err := ctr.writeCount(c)
		if err != nil {
			log.Println(err)
		}
	}
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

func (ctr *Counter) touch(key string) *types.Count {
	if _, ok := ctr.Tmp[key]; !ok {
		ctr.Tmp[key] = &types.Count{
			DashId:   ctr.Config.DashId,
			Hostname: ctr.GetHostname(),
			Logname:  ctr.Logname,
			Keyname:  key,
		}
	}
	return ctr.Tmp[key]
}
func (ctr *Counter) Inc(key string, num float64) *types.Count {
	return ctr.touch(key).Inc(num)
}

func (ctr *Counter) Max(key string, num float64) *types.Count {
	return ctr.touch(key).Max(num)
}

func (ctr *Counter) Min(key string, num float64) *types.Count {
	return ctr.touch(key).Min(num)
}

func (ctr *Counter) Avg(key string, num float64) *types.Count {
	return ctr.touch(key).Avg(num)
}

func (ctr *Counter) Per(key string, taken float64, total float64) *types.Count {
	return ctr.touch(key).Per(taken, total)
}

func (ctr *Counter) Time(key string, d time.Duration) func() time.Duration {
	return ctr.touch(key).Time(d)
}
