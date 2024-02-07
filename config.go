package logr_go_client

import (
	"github.com/504dev/logr-go-client/types"
	"github.com/504dev/logr-go-client/utils"
	"net"
	"os"
	"time"
)

var hostname, _ = os.Hostname()
var pid = os.Getpid()
var commit = utils.ReadGitCommit()
var tag = utils.ReadGitTag()

type Config struct {
	Udp        string
	DashId     int
	PublicKey  string
	PrivateKey string
	Hostname   string
	Version    string
	NoCipher   bool
}

func (c *Config) NewLogger(logname string) (*Logger, error) {
	conn, err := net.Dial("udp", c.Udp)
	if err != nil {
		return nil, err
	}
	res := &Logger{
		Config:  c,
		Logname: logname,
		Prefix:  "{time} {level} ",
		Body:    "[{version}, pid={pid}, {initiator}] {message}",
		Conn:    conn,
		Level:   LevelDebug,
		Console: true,
	}
	res.Counter, err = c.NewCounter(logname)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Config) NewCounterConn(name string, conn net.Conn) *Counter {
	cntr := &Counter{
		Config:  c,
		Logname: name,
		State:   make(map[string]*types.Count),
		Conn:    conn,
	}
	cntr.run(10 * time.Second)
	return cntr
}
func (c *Config) NewCounter(name string) (*Counter, error) {
	conn, err := net.Dial("udp", c.Udp)
	if err != nil {
		return nil, err
	}
	return c.NewCounterConn(name, conn), nil
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
