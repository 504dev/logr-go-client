package types

import (
	"github.com/504dev/logr-go-client/cipher"
	"sync"
	"time"
)

type Count struct {
	sync.RWMutex
	DashId    int    `db:"dash_id"   json:"dash_id,omitempty"`
	Timestamp int64  `db:"timestamp" json:"timestamp"`
	Hostname  string `db:"hostname"  json:"hostname,omitempty"`
	Version   string `db:"version"   json:"version,omitempty"`
	Logname   string `db:"logname"   json:"logname,omitempty"`
	Keyname   string `db:"keyname"   json:"keyname"`
	Metrics   `json:"metrics"`
}

type Metrics struct {
	*Inc
	*Max
	*Min
	*Avg
	*Per
	*Time
}

// for logr usage
func (m Metrics) ToMap() map[string]interface{} {
	res := map[string]interface{}{}
	if m.Inc != nil {
		res["inc"] = m.Inc.Value()
	}
	if m.Max != nil {
		res["max"] = m.Max.Value()
	}
	if m.Min != nil {
		res["min"] = m.Min.Value()
	}
	if m.Avg != nil {
		res["avg"] = m.Avg.Value()
	}
	if m.Per != nil {
		res["per"] = m.Per.Value()
	}
	if m.Time != nil {
		res["time"] = m.Time.Value()
	}
	return res
}

func (c *Count) Decrypt(cipherData []byte, priv string) error {
	c.RLock()
	defer c.RUnlock()
	return cipher.DecodeAesJson(cipherData, priv, c)
}

func (c *Count) Encrypt(priv string) ([]byte, error) {
	c.RLock()
	defer c.RUnlock()
	return cipher.EncryptAesJson(c, priv)
}

func (c *Count) now() {
	c.Timestamp = time.Now().Unix()
}

func (c *Count) Inc(num float64) *Count {
	c.Lock()
	defer c.Unlock()
	if c.Metrics.Inc == nil {
		c.Metrics.Inc = &Inc{}
	}
	c.Metrics.Inc.Val += num - c.Metrics.Inc.Last
	c.now()
	return c
}

func (c *Count) IncLast(num float64) *Count {
	c.Lock()
	defer c.Unlock()
	if c.Metrics.Inc == nil {
		c.Metrics.Inc = &Inc{}
	}
	c.Metrics.Inc.Last = num
	c.now()
	return c
}

func (c *Count) Max(num float64) *Count {
	c.Lock()
	defer c.Unlock()
	if c.Metrics.Max == nil {
		c.Metrics.Max = &Max{Val: num}
	} else if num > c.Metrics.Max.Val {
		c.Metrics.Max.Val = num
	}
	c.now()
	return c
}

func (c *Count) Min(num float64) *Count {
	c.Lock()
	defer c.Unlock()
	if c.Metrics.Min == nil {
		c.Metrics.Min = &Min{Val: num}
	} else if num < c.Metrics.Min.Val {
		c.Metrics.Min.Val = num
	}
	c.now()
	return c
}

func (c *Count) Avg(num float64) *Count {
	c.Lock()
	defer c.Unlock()
	if c.Metrics.Avg == nil {
		c.Metrics.Avg = &Avg{}
	}
	c.Metrics.Avg.Sum += num
	c.Metrics.Avg.Num += 1
	c.now()
	return c
}

func (c *Count) Per(taken float64, total float64) *Count {
	c.Lock()
	defer c.Unlock()
	if c.Metrics.Per == nil {
		c.Metrics.Per = &Per{}
	}
	c.Metrics.Per.Taken += taken
	c.Metrics.Per.Total += total
	c.now()
	return c
}

func (c *Count) Time(duration time.Duration) func() time.Duration {
	c.Lock()
	defer c.Unlock()
	if c.Metrics.Time == nil {
		c.Metrics.Time = &Time{}
	}
	c.Metrics.Time.Duration = duration.Nanoseconds()
	c.now()
	ts := time.Now()
	var delta *time.Duration
	return func() time.Duration {
		if delta == nil {
			tmp := time.Since(ts)
			delta = &tmp
			num := float64(tmp.Nanoseconds()) / float64(duration.Nanoseconds())
			c.Avg(num).Min(num).Max(num)
		}
		return *delta
	}
}

type Inc struct {
	Val  float64 `db:"inc,omitempty" json:"inc,omitempty"`
	Last float64 `db:"-" json:"-"`
}

type Max struct {
	Val float64 `db:"max,omitempty" json:"max,omitempty"`
}

type Min struct {
	Val float64 `db:"min,omitempty" json:"min,omitempty"`
}

type Avg struct {
	Sum float64 `db:"avg_sum" json:"avg_sum,omitempty"`
	Num int     `db:"avg_num" json:"avg_num,omitempty"`
}

type Per struct {
	Taken float64 `db:"per_tkn" json:"per_tkn,omitempty"`
	Total float64 `db:"per_ttl" json:"per_ttl,omitempty"`
}

type Time struct {
	Duration int64 `db:"time_dur" json:"time_dur,omitempty"`
}

func (i *Inc) Value() float64 {
	return i.Val
}

func (m *Max) Value() float64 {
	return m.Val
}

func (m *Min) Value() float64 {
	return m.Val
}

func (a *Avg) Value() float64 {
	return a.Sum / float64(a.Num)
}

func (p *Per) Value() float64 {
	return p.Taken / p.Total * 100
}

func (t *Time) Value() int64 {
	return t.Duration
}
