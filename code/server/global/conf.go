package global

import (
	"os"

	"github.com/lwch/runtime"
	"gopkg.in/yaml.v2"
)

type Configure struct {
	Listen uint16
	Secret string
	TLSKey string
	TLSCrt string
}

func LoadConf(dir string) *Configure {
	var cfg struct {
		Listen uint16 `yaml:"listen"`
		Secret string `yaml:"secret"`
		TLS    struct {
			Key string `yaml:"key"`
			Crt string `yaml:"crt"`
		} `yaml:"tls"`
	}
	f, err := os.Open(dir)
	runtime.Assert(err)
	defer f.Close()
	runtime.Assert(yaml.NewDecoder(f).Decode(&cfg))
	return &Configure{
		Listen: cfg.Listen,
		Secret: cfg.Secret,
		TLSKey: cfg.TLS.Key,
		TLSCrt: cfg.TLS.Crt,
	}
}
