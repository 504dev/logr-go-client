package go_kidlog

type Config struct {
	Udp        string
	DashId     int
	PublicKey  string
	PrivateKey string
	Hostname   string
}

func (c *Config) GetHostname() string {
	if c.Hostname != "" {
		return c.Hostname
	}
	return hostname
}
