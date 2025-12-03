package config

import "time"

type proxyCfg struct {
	UpstreamURL         string        `toml:"upstreamURL"`
	MaxIdleConns        int           `toml:"maxIdleConn"`
	MaxIdleConnsPerHost int           `toml:"maxIdleConnPerHost"`
	IdleConnTimeout     time.Duration `toml:"idleConnTimeout"`
}

type SystemCfg struct {
	ListenAddr string   `toml:"listenaddr"`
	ProxyCfg      proxyCfg `toml:"proxy"`
	CacheCfg      cacheCfg `toml:"cache"`
}

type cacheCfg struct {
	Enabled bool `toml:"enabled"`
	CacheCapacity int `toml:"cacheCapacity"`
	DefaultTTL int `toml:"defaultTTL"`
}
