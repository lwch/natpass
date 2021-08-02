package global

import (
	"os"

	"github.com/lwch/runtime"
	"gopkg.in/yaml.v2"
)

type Tunnel struct {
	Name       string `yaml:"name"`
	Target     string `yaml:"target"`
	Type       string `yaml:"type"`
	LocalAddr  string `yaml:"local_addr"`
	LocalPort  uint16 `yaml:"local_port"`
	RemoteAddr string `yaml:"remote_addr"`
	RemotePort uint16 `yaml:"remote_port"`
}

type Configure struct {
	ID      string
	Server  string
	Secret  string
	Tunnels []Tunnel
}

func LoadConf(dir string) *Configure {
	var cfg struct {
		ID     string   `yaml:"id"`
		Server string   `yaml:"server"`
		Secret string   `yaml:"secret"`
		Tunnel []Tunnel `yaml:"tunnel"`
	}
	f, err := os.Open(dir)
	runtime.Assert(err)
	defer f.Close()
	runtime.Assert(yaml.NewDecoder(f).Decode(&cfg))
	for i, t := range cfg.Tunnel {
		if t.Type != "tcp" {
			t.Type = "udp"
		}
		cfg.Tunnel[i] = t
	}
	return &Configure{
		ID:      cfg.ID,
		Server:  cfg.Server,
		Secret:  cfg.Secret,
		Tunnels: cfg.Tunnel,
	}
}
