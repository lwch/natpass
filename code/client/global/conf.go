package global

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lwch/natpass/code/hash"
	"github.com/lwch/natpass/code/utils"
	"github.com/lwch/runtime"
	"github.com/lwch/yaml"
)

// Rule rule config
type Rule struct {
	Name      string `yaml:"name"`
	Target    string `yaml:"target"`
	Type      string `yaml:"type"`
	LocalAddr string `yaml:"local_addr"`
	LocalPort uint16 `yaml:"local_port"`
	// shell
	Exec string   `yaml:"exec"`
	Env  []string `yaml:"env"`
	// vnc
	Fps uint32 `yaml:"fps"`
}

// Configure client configure
type Configure struct {
	ID               string
	Server           string
	UseSSL           bool
	SSLInsecure      bool
	Hasher           *hash.Hasher
	Links            int
	LogDir           string
	LogSize          utils.Bytes
	LogRotate        int
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	DashboardEnabled bool
	DashboardListen  string
	DashboardPort    uint16
	Rules            []*Rule
	CodeDir          string
}

// LoadConf load configure file
func LoadConf(dir string) *Configure {
	var cfg struct {
		ID     string `yaml:"id"`
		Server string `yaml:"server"`
		Secret string `yaml:"secret"`
		SSL    struct {
			Enabled  bool `yaml:"enabled"`
			Insecure bool `yaml:"insecure"`
		} `yaml:"ssl"`
		Link struct {
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
		Rules   []*Rule `yaml:"rules"`
		CodeDir string  `yaml:"codedir"`
	}
	cfg.ID = "unset"
	cfg.Server = "127.0.0.1:6154"
	cfg.SSL.Enabled = false
	cfg.SSL.Insecure = false
	cfg.Dashboard.Enabled = true
	cfg.Dashboard.Listen = "0.0.0.0"
	cfg.Dashboard.Port = 8080
	cfg.Secret = "0123456789"
	cfg.Link.ReadTimeout = time.Second
	cfg.Link.WriteTimeout = time.Second
	cfg.Log.Dir = "./logs"
	cfg.Log.Size = 50 * 1024 * 1024
	cfg.Log.Rotate = 7
	cfg.CodeDir = "./code"
	runtime.Assert(yaml.Decode(dir, &cfg))
	for i, t := range cfg.Rules {
		switch t.Type {
		case "shell", "vnc", "bench", "code-server":
		default:
			panic(fmt.Sprintf("unsupported type: %s", t.Type))
		}
		cfg.Rules[i] = t
	}
	if cfg.Link.ReadTimeout <= 0 {
		cfg.Link.ReadTimeout = 5 * time.Second
	}
	if cfg.Link.WriteTimeout <= 0 {
		cfg.Link.WriteTimeout = 5 * time.Second
	}
	if !filepath.IsAbs(cfg.Log.Dir) {
		dir, err := os.Executable()
		runtime.Assert(err)
		cfg.Log.Dir = filepath.Join(filepath.Dir(dir), cfg.Log.Dir)
	}
	if !filepath.IsAbs(cfg.CodeDir) {
		dir, err := os.Executable()
		runtime.Assert(err)
		cfg.CodeDir = filepath.Join(filepath.Dir(dir), cfg.CodeDir)
	}
	return &Configure{
		ID:               cfg.ID,
		Server:           cfg.Server,
		UseSSL:           cfg.SSL.Enabled,
		SSLInsecure:      cfg.SSL.Insecure,
		Hasher:           hash.New(cfg.Secret, 60),
		ReadTimeout:      cfg.Link.ReadTimeout,
		WriteTimeout:     cfg.Link.WriteTimeout,
		LogDir:           cfg.Log.Dir,
		LogSize:          cfg.Log.Size,
		LogRotate:        cfg.Log.Rotate,
		DashboardEnabled: cfg.Dashboard.Enabled,
		DashboardListen:  cfg.Dashboard.Listen,
		DashboardPort:    cfg.Dashboard.Port,
		Rules:            cfg.Rules,
		CodeDir:          cfg.CodeDir,
	}
}
