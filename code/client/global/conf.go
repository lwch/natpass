package global

import (
	"crypto/md5"
	"natpass/code/utils"
	"time"

	"github.com/lwch/runtime"
	"github.com/lwch/yaml"
)

// Tunnel tunnel config
type Tunnel struct {
	Name       string   `yaml:"name"`
	Target     string   `yaml:"target"`
	Type       string   `yaml:"type"`
	LocalAddr  string   `yaml:"local_addr"`
	LocalPort  uint16   `yaml:"local_port"`
	RemoteAddr string   `yaml:"remote_addr"`
	RemotePort uint16   `yaml:"remote_port"`
	Exec       string   `yaml:"exec"`
	Env        []string `yaml:"env"`
}

// Configure client configure
type Configure struct {
	ID               string
	Server           string
	Enc              [md5.Size]byte
	Links            int
	LogDir           string
	LogSize          utils.Bytes
	LogRotate        int
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	DashboardEnabled bool
	DashboardListen  string
	DashboardPort    uint16
	Tunnels          []Tunnel
}

// LoadConf load configure file
func LoadConf(dir string) *Configure {
	var cfg struct {
		ID     string `yaml:"id"`
		Server string `yaml:"server"`
		Secret string `yaml:"secret"`
		Link   struct {
			Connections  int           `yaml:"connections"`
			ReadTimeout  time.Duration `yaml:"read_timeout"`
			WriteTimeout time.Duration `yaml:"write_timeout"`
		} `yaml:"link"`
		Log struct {
			Dir    string      `yaml:"dir"`
			Size   utils.Bytes `yaml:"size"`
			Rotate int         `yaml:"rotate"`
		} `yaml:"log"`
		Dashboard struct {
			Enabled bool   `yaml:"enabled"`
			Listen  string `yaml:"listen"`
			Port    uint16 `yaml:"port"`
		} `yaml:"dashboard"`
		Tunnel []Tunnel `yaml:"tunnel"`
	}
	runtime.Assert(yaml.Decode(dir, &cfg))
	for i, t := range cfg.Tunnel {
		switch t.Type {
		case "tcp", "shell":
		default:
			t.Type = "udp"
		}
		cfg.Tunnel[i] = t
	}
	if cfg.Link.Connections <= 0 {
		cfg.Link.Connections = 3
	}
	if cfg.Link.ReadTimeout <= 0 {
		cfg.Link.ReadTimeout = 5 * time.Second
	}
	if cfg.Link.WriteTimeout <= 0 {
		cfg.Link.WriteTimeout = 5 * time.Second
	}
	return &Configure{
		ID:               cfg.ID,
		Server:           cfg.Server,
		Enc:              md5.Sum([]byte(cfg.Secret)),
		Links:            cfg.Link.Connections,
		ReadTimeout:      cfg.Link.ReadTimeout,
		WriteTimeout:     cfg.Link.WriteTimeout,
		LogDir:           cfg.Log.Dir,
		LogSize:          cfg.Log.Size,
		LogRotate:        cfg.Log.Rotate,
		DashboardEnabled: cfg.Dashboard.Enabled,
		DashboardListen:  cfg.Dashboard.Listen,
		DashboardPort:    cfg.Dashboard.Port,
		Tunnels:          cfg.Tunnel,
	}
}
