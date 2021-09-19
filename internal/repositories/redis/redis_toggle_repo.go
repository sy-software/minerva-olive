package redis

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/sy-software/minerva-olive/internal/core/domain"
)

const (
	FlagPrefix string = "flags:"

	StatusKey string = "status"
	DataKey   string = "data"
)

type RedisToggleRepo struct {
	db *RedisDB
}

func NewRedisToggleRepo(config *domain.Config, db *RedisDB) *RedisToggleRepo {
	return &RedisToggleRepo{
		db: db,
	}
}

func (repo *RedisToggleRepo) GetFlag(name string, ctx context.Context) domain.ToggleFlag {
	return repo.GetFlagWithDefaults(name, false, nil, ctx)
}

func (repo *RedisToggleRepo) GetFlagWithDefaults(name string, defStatus bool, defData interface{}, ctx context.Context) domain.ToggleFlag {
	cmd := repo.db.Client.HGetAll(context.Background(), FlagPrefix+name)
	if cmd.Err() != nil {
		return domain.ToggleFlag{
			Status: defStatus,
			Data:   defData,
		}
	}

	status := false
	rawStatus, ok := cmd.Val()[StatusKey]
	log.Debug().Msgf("Toggle flag raw status: %v", rawStatus)
	if !ok {
		status = defStatus
	} else {
		status = rawStatus == "1"
	}

	return domain.ToggleFlag{
		Status: status,
	}
}

func (repo *RedisToggleRepo) SetFlag(name string, status bool, data interface{}) error {
	cmd := repo.db.Client.HSet(context.Background(), FlagPrefix+name, StatusKey, status, DataKey, data)
	return cmd.Err()
}
