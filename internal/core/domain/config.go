package domain

import (
	"encoding/json"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	DEFAULT_CONFIG_FILE = "./config.json"
	CONFIG_FILE_VAR     = "CONFIG_FILE"
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
	// Server bind IP default 0.0.0.0
	Host string `json:"host,omitempty"`
	// Server bind port default 8080
	Port int `json:"port,omitempty"`
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
		Host:     "127.0.0.1",
		Port:     8080,
	}
}

// LoadConfiguration Loads the configuration object from a json file
func LoadConfigurationFile(file string) Config {
	config := DefaultConfig()
	configFile, err := os.Open(file)

	if err != nil {
		log.Warn().Err(err).Msg("Can't load config file. Default values will be used instead")
	}

	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func LoadConfig() Config {
	log.Info().Msg("Loading configuration")
	config := DefaultConfig()

	configFile := os.Getenv(CONFIG_FILE_VAR)
	if configFile == "" {
		configFile = DEFAULT_CONFIG_FILE
	}

	log.Info().Msgf("Looking for configuration from: %s", configFile)
	config = LoadConfigurationFile(configFile)
	log.Info().Msg("Configuration loaded")
	return config
}
