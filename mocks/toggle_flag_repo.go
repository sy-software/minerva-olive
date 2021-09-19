package mocks

import (
	"context"

	"github.com/sy-software/minerva-olive/internal/core/domain"
)

type ToggleFlagRepo struct {
	Flags map[string]domain.ToggleFlag
}

func NewToggleFlagRepo(flags map[string]domain.ToggleFlag) *ToggleFlagRepo {
	return &ToggleFlagRepo{
		Flags: flags,
	}
}

func (repo *ToggleFlagRepo) GetFlag(name string, ctx context.Context) domain.ToggleFlag {
	return repo.GetFlagWithDefaults(name, false, nil, ctx)
}

func (repo *ToggleFlagRepo) GetFlagWithDefaults(name string, defStatus bool, defData interface{}, ctx context.Context) domain.ToggleFlag {
	f, ok := repo.Flags[name]
	if !ok {
		return domain.ToggleFlag{
			Status: defStatus,
			Data:   defData,
		}
	}
	return f
}

func (repo *ToggleFlagRepo) SetFlag(name string, status bool, data interface{}) error {
	repo.Flags[name] = domain.ToggleFlag{
		Status: status,
		Data:   data,
	}

	return nil
}
