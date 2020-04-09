package go_kidlog

import "os"

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
