package global

import (
	"crypto/md5"
	"natpass/code/utils"
	"os"

	"github.com/lwch/runtime"
	"gopkg.in/yaml.v2"
)

// Tunnel tunnel config
type Tunnel struct {
	Name       string `yaml:"name"`
	Target     string `yaml:"target"`
	Type       string `yaml:"type"`
	LocalAddr  string `yaml:"local_addr"`
	LocalPort  uint16 `yaml:"local_port"`
	RemoteAddr string `yaml:"remote_addr"`
	RemotePort uint16 `yaml:"remote_port"`
}

// Configure client configure
type Configure struct {
	ID        string
	Server    string
	Enc       [md5.Size]byte
	LogDir    string
	LogSize   utils.Bytes
	LogRotate int
	Tunnels   []Tunnel
}

// LoadConf load configure file
func LoadConf(dir string) *Configure {
	var cfg struct {
		ID     string `yaml:"id"`
		Server string `yaml:"server"`
		Secret string `yaml:"secret"`
		Log    struct {
			Dir    string      `yaml:"dir"`
			Size   utils.Bytes `yaml:"size"`
			Rotate int         `yaml:"rotate"`
		} `yaml:"log"`
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
		ID:        cfg.ID,
		Server:    cfg.Server,
		Enc:       md5.Sum([]byte(cfg.Secret)),
		LogDir:    cfg.Log.Dir,
		LogSize:   cfg.Log.Size,
		LogRotate: cfg.Log.Rotate,
		Tunnels:   cfg.Tunnel,
	}
}
