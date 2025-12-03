package config

import (
	"flag"
	"log"
	"time"

	"github.com/BurntSushi/toml"
)

var defaultSystemCfg = &SystemCfg{
	ListenAddr: ":8000",
	Proxy: proxyCfg{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     10 * time.Second,
	},
}

func LoadConfig() (*SystemCfg, error) {
	configFile := flag.String("config", "config.toml", "location of config file")
	flag.Parse()
	config := defaultSystemCfg

	if _, err := toml.DecodeFile(*configFile, config); err != nil {
		log.Fatal(err)
	}

	return config, nil
}
