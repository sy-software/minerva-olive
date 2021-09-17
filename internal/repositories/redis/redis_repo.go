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
	CfgSetNames  string = "set:names"
)

type RedisRepo struct {
	db *RedisDB
}

func NewRedisRepo(config *domain.Config, db *RedisDB) *RedisRepo {
	return &RedisRepo{
		db: db,
	}
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

func (repo *RedisRepo) RemoveJSON(key string) error {
	ctx := context.Background()
	cmds, err := repo.db.Client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		zrem := p.ZRem(ctx, AgeTracker, key)
		if zrem.Err() != nil {
			return zrem.Err()
		}

		statusDel := p.Del(ctx, JSONPrefix+key)

		if statusDel.Err() != nil {
			return statusDel.Err()
		}

		return nil
	})

	log.Info().Msgf("CACHE: Remove JSON commands: %+v", cmds)
	return err
}

func (repo *RedisRepo) CreateSet(set domain.ConfigSet) (domain.ConfigSet, error) {
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

	cmds, err := repo.db.Client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		cmdSet := repo.db.Client.Set(ctx, key, jsonBytes, 0)

		if cmdSet.Err() != nil {
			return cmdSet.Err()
		}
		cmdName := p.ZAdd(ctx, CfgSetNames, &redis.Z{
			Score:  float64(time.Now().UTC().UnixNano()),
			Member: key,
		})

		if cmdName.Err() != nil {
			return cmdName.Err()
		}

		return nil
	})

	log.Info().Msgf("CACHE: Create Set commands: %+v", cmds)
	return set, err
}

func (repo *RedisRepo) GetSet(name string) (*domain.ConfigSet, error) {
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
	ctx := context.Background()
	start := skip
	end := skip + limit - 1
	cmd := repo.db.Client.ZRange(ctx, CfgSetNames, int64(start), int64(end))

	if cmd.Err() != nil {
		if cmd.Err() == redis.Nil {
			return []string{}, nil
		}

		return nil, cmd.Err()
	}
	return cmd.Val(), nil
}

func (repo *RedisRepo) DeleteSet(name string) (domain.ConfigSet, error) {
	ctx := context.Background()
	key := CfgSetPrefix + name
	exists := repo.db.Client.Exists(ctx, key)
	if exists.Val() == 0 {
		return domain.ConfigSet{}, ports.ErrConfigNotExists
	}

	cmd := repo.db.Client.Get(ctx, CfgSetPrefix+name)
	if cmd.Err() != nil {
		if cmd.Err() == redis.Nil {
			return domain.ConfigSet{}, ports.ErrConfigNotExists
		}
		return domain.ConfigSet{}, cmd.Err()
	}

	var set domain.ConfigSet
	json.Unmarshal([]byte(cmd.Val()), &set)

	cmds, err := repo.db.Client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		cmdDel := repo.db.Client.Del(ctx, key)

		if cmdDel.Err() != nil {
			return cmdDel.Err()
		}
		cmdName := p.ZRem(ctx, CfgSetNames, key)

		if cmdName.Err() != nil {
			return cmdName.Err()
		}

		return nil
	})

	log.Info().Msgf("CACHE: Delete Set commands: %+v", cmds)
	return set, err
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
	jsonBytes, err := json.Marshal(set)
	if err != nil {
		return domain.ConfigSet{}, err
	}
	setCmd := repo.db.Client.Set(ctx, CfgSetPrefix+setName, jsonBytes, redis.KeepTTL)
	if setCmd.Err() != nil {
		return set, setCmd.Err()
	}

	return set, nil
}

func (repo *RedisRepo) UpdateItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
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

	_, err = set.Update(item)
	if err != nil {
		return domain.ConfigSet{}, err
	}

	set.UpdateDate = datetime.UnixUTCNow()
	jsonBytes, err := json.Marshal(set)
	if err != nil {
		return domain.ConfigSet{}, err
	}
	setCmd := repo.db.Client.Set(ctx, CfgSetPrefix+setName, jsonBytes, redis.KeepTTL)
	if setCmd.Err() != nil {
		return set, setCmd.Err()
	}

	return set, nil
}

func (repo *RedisRepo) RemoveItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
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

	_, err = set.Delete(item.Key)
	if err != nil {
		return domain.ConfigSet{}, err
	}

	set.UpdateDate = datetime.UnixUTCNow()
	jsonBytes, err := json.Marshal(set)
	if err != nil {
		return domain.ConfigSet{}, err
	}
	setCmd := repo.db.Client.Set(ctx, CfgSetPrefix+setName, jsonBytes, redis.KeepTTL)
	if setCmd.Err() != nil {
		return set, setCmd.Err()
	}

	return set, nil
}
