package config

import (
	"flag"
	"log"
	"time"

	"github.com/BurntSushi/toml"
)

var SystemConfig *SystemCfg

func NewSystemCfg() *SystemCfg {
	return &SystemCfg{
		ListenAddr: ":8000",
		Proxy: proxyCfg{
			UpstreamURL:         "http://localhost:9000/api",
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     10 * time.Second,
		},
	}
}

func LoadConfig() {
	configFile := flag.String("config", "config.toml", "location of config file")
	flag.Parse()
	config := NewSystemCfg()

	if _, err := toml.DecodeFile(*configFile, config); err != nil {
		log.Fatal(err)
	}

	SystemConfig = config
}


