package ports

import (
	"context"
	"errors"

	"github.com/sy-software/minerva-olive/internal/core/domain"
)

// Errors
var (
	ErrSecretNoExists   = errors.New("secret does not exists")
	ErrDuplicatedConfig = errors.New("config set already exists")
	ErrConfigNotExists  = errors.New("config does not exists")
	ErrOldValue         = errors.New("cached value is older than expected")
)

type Repo interface {
	CreateSet(set domain.ConfigSet) (domain.ConfigSet, error)
	GetSet(name string) (*domain.ConfigSet, error)
	GetSetNames(limit int, skip int) ([]string, error)
	DeleteSet(name string) (domain.ConfigSet, error)
	AddItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	UpdateItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	RemoveItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
}

type CacheRepo interface {
	SaveJSON(json []byte, key string, ttl int) error
	GetJSON(key string, maxAge int) ([]byte, error)
	RemoveJSON(key string) error
}

type Secret interface {
	Get(name string) (string, error)
}

type Notifier interface {
	SetUpdated(name string, set domain.ConfigSet) error
}

type ToggleRepo interface {
	GetFlag(name string, ctx context.Context) domain.ToggleFlag
	GetFlagWithDefaults(name string, defStatus bool, defData interface{}, ctx context.Context) domain.ToggleFlag
	SetFlag(name string, status bool, data interface{}) error
}
