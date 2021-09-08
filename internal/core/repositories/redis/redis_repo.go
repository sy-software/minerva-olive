package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"github.com/sy-software/minerva-go-utils/datetime"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
)

const (
	AgeTracker   string = "json:age"
	JSONPrefix   string = "json:"
	CfgSetPrefix string = "set:"
)

type RedisRepo struct {
	db *RedisDB
}

func NewRedisRepo(config *domain.Config) (*RedisRepo, error) {
	db, err := GetRedisDB(config)
	if err != nil {
		return nil, err
	}

	return &RedisRepo{
		db: db,
	}, nil
}

func (repo *RedisRepo) SaveJSON(json []byte, key string, ttl int) error {
	redisTTL := ttl
	if ttl != domain.InfiniteTTL {
		redisTTL = 0
	}
	ctx := context.Background()
	cmds, err := repo.db.Client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		statusAge := p.ZAdd(ctx, AgeTracker, &redis.Z{
			Score:  float64(datetime.UnixUTCNow().UnixNano()),
			Member: key,
		})

		if statusAge.Err() != nil {
			return statusAge.Err()
		}
		statusSave := p.Set(ctx, JSONPrefix+key, json, time.Duration(redisTTL))

		if statusSave.Err() != nil {
			return statusSave.Err()
		}

		return nil
	})

	log.Info().Msgf("CACHE: Save JSON commands: %+v", cmds)
	return err
}

func (repo *RedisRepo) GetJSON(key string, maxAge int) ([]byte, error) {
	ctx := context.Background()

	if maxAge != domain.AnyAge {
		ageCmd := repo.db.Client.ZScore(ctx, AgeTracker, key)
		// TODO: Should we continue if the value has no age?
		if ageCmd.Err() != nil {
			if ageCmd.Err() == redis.Nil {
				return nil, ports.ErrConfigNotExists
			}
			return nil, ageCmd.Err()
		}
		now := datetime.UnixUTCNow()
		age := time.Unix(0, int64(ageCmd.Val()))

		if age.Add(time.Duration(maxAge)).Before(now) {
			return nil, ports.ErrOldValue
		}
	}

	valCmd := repo.db.Client.Get(ctx, JSONPrefix+key)
	if valCmd.Err() != nil {
		if valCmd.Err() == redis.Nil {
			return nil, ports.ErrConfigNotExists
		}
		return nil, valCmd.Err()
	}

	return []byte(valCmd.Val()), nil
}

func (repo *RedisRepo) CreateSet(set domain.ConfigSet, ttl int) (domain.ConfigSet, error) {
	ctx := context.Background()
	key := CfgSetPrefix + set.Name
	exists := repo.db.Client.Exists(ctx, key)
	if exists.Val() == 1 {
		return set, ports.ErrDuplicatedConfig
	}

	jsonBytes, err := json.Marshal(set)
	if err != nil {
		return set, err
	}
	cmd := repo.db.Client.Set(ctx, key, jsonBytes, 0)
	if cmd.Err() != nil {
		if cmd.Err() == redis.Nil {
			return set, ports.ErrConfigNotExists
		}
		return set, cmd.Err()
	}

	return set, nil
}

func (repo *RedisRepo) GetSet(name string, maxAge int) (*domain.ConfigSet, error) {
	ctx := context.Background()
	cmd := repo.db.Client.Get(ctx, CfgSetPrefix+name)
	if cmd.Err() != nil {
		if cmd.Err() == redis.Nil {
			return nil, ports.ErrConfigNotExists
		}
		return nil, cmd.Err()
	}

	var set domain.ConfigSet
	json.Unmarshal([]byte(cmd.Val()), &set)
	return &set, nil
}

func (repo *RedisRepo) GetSetNames(limit int, skip int) ([]string, error) {
	panic("not implemented")
}

func (repo *RedisRepo) DeleteSet(name string) (domain.ConfigSet, error) {
	panic("not implemented")
}

func (repo *RedisRepo) AddItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
	ctx := context.Background()
	cmd := repo.db.Client.Get(ctx, CfgSetPrefix+setName)
	if cmd.Err() != nil {
		if cmd.Err() == redis.Nil {
			return domain.ConfigSet{}, ports.ErrConfigNotExists
		}
		return domain.ConfigSet{}, cmd.Err()
	}

	var set domain.ConfigSet
	err := json.Unmarshal([]byte(cmd.Val()), &set)
	if err != nil {
		return domain.ConfigSet{}, err
	}

	err = set.Add(item)
	if err != nil {
		return domain.ConfigSet{}, err
	}

	set.UpdateDate = datetime.UnixUTCNow()
	setCmd := repo.db.Client.Set(ctx, CfgSetPrefix+setName, set, redis.KeepTTL)
	if setCmd.Err() != nil {
		return set, cmd.Err()
	}

	return set, nil
}

func (repo *RedisRepo) UpdateItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
	panic("not implemented")
}

func (repo *RedisRepo) RemoveItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
	panic("not implemented")
}
