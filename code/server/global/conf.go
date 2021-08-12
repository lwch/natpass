package global

import (
	"crypto/md5"
	"os"

	"github.com/lwch/runtime"
	"gopkg.in/yaml.v2"
)

type Configure struct {
	Listen    uint16
	Enc       [md5.Size]byte
	TLSKey    string
	TLSCrt    string
	LogDir    string
	LogSize   int
	LogRotate int
}

func LoadConf(dir string) *Configure {
	var cfg struct {
		Listen uint16 `yaml:"listen"`
		Secret string `yaml:"secret"`
		Log    struct {
			Dir    string `yaml:"dir"`
			Size   int    `yaml:"size"`
			Rotate int    `yaml:"rotate"`
		} `yaml:"log"`
		TLS struct {
			Key string `yaml:"key"`
			Crt string `yaml:"crt"`
		} `yaml:"tls"`
	}
	f, err := os.Open(dir)
	runtime.Assert(err)
	defer f.Close()
	runtime.Assert(yaml.NewDecoder(f).Decode(&cfg))
	return &Configure{
		Listen:    cfg.Listen,
		Enc:       md5.Sum([]byte(cfg.Secret)),
		TLSKey:    cfg.TLS.Key,
		TLSCrt:    cfg.TLS.Crt,
		LogDir:    cfg.Log.Dir,
		LogSize:   cfg.Log.Size,
		LogRotate: cfg.Log.Rotate,
	}
}
