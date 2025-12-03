package config

import "time"

type proxyCfg struct {
	UpstreamURL string `toml:"upstreamURL"`
	MaxIdleConns int `toml:"maxIdleConn"`
	MaxIdleConnsPerHost int `toml:"maxIdleConnPerHost"`
	IdleConnTimeout time.Duration `toml:"idleConnTimeout"`
}

type SystemCfg struct {
	ListenAddr string `toml:"listenaddr"`
	Proxy proxyCfg `toml:"proxy"`
}