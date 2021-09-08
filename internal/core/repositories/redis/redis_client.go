package redis

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"github.com/sy-software/minerva-olive/internal/core/domain"
)

var (
	ErrCantConnect = errors.New("can't connect to redis")
)

type RedisDB struct {
	config *domain.Config
	Client *redis.Client
}

var rdbInstance *RedisDB
var rdbOnce sync.Once

// GetRedisDB creates or returns a singleton connection to a RedisDB instance
func GetRedisDB(config *domain.Config) (*RedisDB, error) {
	var dbErr error

	rdbOnce.Do(func() {
		log.Info().Msg("Initializing Redis DB connection")
		addr := fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port)
		opts := &redis.Options{}
		opts.Addr = addr
		if len(config.Redis.Username) > 0 {
			opts.Username = config.Redis.Username
			opts.Password = config.Redis.Password
		}

		opts.DB = config.Redis.DB
		opts.MaxRetries = config.Redis.MaxRetries
		opts.DialTimeout = config.Redis.ConnectionTimeout
		opts.ReadTimeout = config.Redis.ReadTimeout
		opts.WriteTimeout = config.Redis.WriteTimeout
		opts.PoolSize = config.Redis.PoolSize

		rdb := redis.NewClient(opts)

		if rdb == nil {
			dbErr = ErrCantConnect
		}

		rdbInstance = &RedisDB{
			config: config,
			Client: rdb,
		}

		log.Info().Msg("DB Connected")
	})

	return rdbInstance, dbErr
}
