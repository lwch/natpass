package global

import (
	"os"
	"path/filepath"
	"time"

	"github.com/lwch/natpass/code/hash"
	"github.com/lwch/natpass/code/utils"
	"github.com/lwch/runtime"
	"github.com/lwch/yaml"
)

// Configure server configure
type Configure struct {
	Listen       uint16
	Hasher       *hash.Hasher
	TLSKey       string
	TLSCrt       string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	LogDir       string
	LogSize      utils.Bytes
	LogRotate    int
}

// LoadConf load configure file
func LoadConf(dir string) *Configure {
	var cfg struct {
		Listen uint16 `yaml:"listen"`
		Secret string `yaml:"secret"`
		Link   struct {
			ReadTimeout  time.Duration `yaml:"read_timeout"`
			WriteTimeout time.Duration `yaml:"write_timeout"`
		} `yaml:"link"`
		Log struct {
			Dir    string      `yaml:"dir"`
			Size   utils.Bytes `yaml:"size"`
			Rotate int         `yaml:"rotate"`
		} `yaml:"log"`
		TLS struct {
			Key string `yaml:"key"`
			Crt string `yaml:"crt"`
		} `yaml:"tls"`
	}
	cfg.Listen = 6154
	cfg.Secret = "0123456789"
	cfg.Link.ReadTimeout = time.Second
	cfg.Link.WriteTimeout = time.Second
	cfg.Log.Dir = "./logs"
	cfg.Log.Size = 50 * 1024 * 1024
	cfg.Log.Rotate = 7
	runtime.Assert(yaml.Decode(dir, &cfg))
	if !filepath.IsAbs(cfg.Log.Dir) {
		dir, err := os.Executable()
		runtime.Assert(err)
		cfg.Log.Dir = filepath.Join(filepath.Dir(dir), cfg.Log.Dir)
	}
	return &Configure{
		Listen:       cfg.Listen,
		Hasher:       hash.New(cfg.Secret, 60),
		TLSKey:       cfg.TLS.Key,
		TLSCrt:       cfg.TLS.Crt,
		ReadTimeout:  cfg.Link.ReadTimeout,
		WriteTimeout: cfg.Link.WriteTimeout,
		LogDir:       cfg.Log.Dir,
		LogSize:      cfg.Log.Size,
		LogRotate:    cfg.Log.Rotate,
	}
}
