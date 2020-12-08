package logr_go_client

import (
	"github.com/504dev/logr/types"
	"net"
	"os"
	"time"
)

var hostname, _ = os.Hostname()
var pid = os.Getpid()
var commit = readCommit()
var tag = readTag()

type Config struct {
	Udp        string
	DashId     int
	PublicKey  string
	PrivateKey string
	Hostname   string
	Version    string
}

func (c *Config) NewLogger(logname string) (*Logger, error) {
	conn, err := net.Dial("udp", c.Udp)
	res := &Logger{
		Config:  c,
		Logname: logname,
		Prefix:  "{time} {level} ",
		Body:    "[{version}, pid={pid}, {initiator}] {message}",
		Conn:    conn,
		Level:   LevelDebug,
		Console: true,
	}
	res.Counter, _ = c.NewCounter(logname)
	return res, err
}

func (c *Config) NewCounter(name string) (*Counter, error) {
	cntr := &Counter{
		Config:  c,
		Logname: name,
		Tmp:     make(map[string]*types.Count),
	}
	err := cntr.connect()
	cntr.run(10 * time.Second)
	return cntr, err
}

func (c *Config) DefaultSystemCounter() (*Counter, error) {
	cntr, err := c.NewCounter("system.log")
	if err != nil {
		return nil, err
	}
	cntr.WatchSystem()
	return cntr, nil
}

func (c *Config) DefaultProcessCounter() (*Counter, error) {
	cntr, err := c.NewCounter("process.log")
	if err != nil {
		return nil, err
	}
	cntr.WatchProcess()
	return cntr, nil
}

func (c *Config) GetHostname() string {
	if c.Hostname != "" {
		return c.Hostname
	}
	return hostname
}

func (c *Config) GetPid() int {
	return pid
}

func (c *Config) GetVersion() string {
	if c.Version != "" {
		return c.Version
	} else if tag != "" {
		return tag
	} else if len(commit) >= 6 {
		return commit[:6]
	} else {
		return ""
	}
}
