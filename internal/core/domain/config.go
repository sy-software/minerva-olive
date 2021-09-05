package domain

import (
	"time"
)

type RedisCfg struct {
	// Database host, default: 127.0.0.1
	Host string `json:"host,omitempty"`
	// Database port, default: 6379
	Port int `json:"port,omitempty"`
	// Database username. Omit if the DB have no authentication
	Username string `json:"username,omitempty"`
	// Database password. Omit if the DB have no authentication
	Password string `json:"password,omitempty"`
	// Database selected, default: 1
	DB int `json:"db,omitempty"`
	// Maximum number of retries before giving up, default: 3
	MaxRetries int `json:"maxRetries,omitempty"`
	// Database connection timeout, default: 10 seconds
	ConnectionTimeout time.Duration `json:"connectionTimeout,omitempty"`
	// Database read timeout, default: 1s
	ReadTimeout time.Duration `json:"readTimeout,omitempty"`
	// Database read timeout, default: 1s
	WriteTimeout time.Duration `json:"writeTimeout,omitempty"`
	// Max number of socket connections, default: 10
	PoolSize int `json:"poolSize,omitempty"`
}

// Config is used to load this service own config
type Config struct {
	// Redis connection configurations
	Redis RedisCfg `json:"redisConfig"`
	// TTL for stored cache, default infinite
	CacheTTL time.Duration
}

// DefaultConfig returns a configuration object with the default values
func DefaultConfig() Config {
	return Config{
		Redis: RedisCfg{
			Host:              "127.0.0.1",
			Port:              6379,
			DB:                1,
			MaxRetries:        3,
			ConnectionTimeout: time.Duration(10) * time.Second,
			ReadTimeout:       time.Duration(1) * time.Second,
			WriteTimeout:      time.Duration(1) * time.Second,
			PoolSize:          10,
		},
		CacheTTL: time.Duration(InfiniteTTL),
	}
}
