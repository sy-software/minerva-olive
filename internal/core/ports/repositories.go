package ports

import (
	"errors"

	"github.com/sy-software/minerva-olive/internal/core/domain"
)

// Utility const
const (
	InfiniteTTL = -1
	AnyAge      = -1
)

// Errors
var (
	ErrSecretNoExists   = errors.New("secret does not exists")
	ErrDuplicatedConfig = errors.New("config set already exists")
	ErrConfigNotExists  = errors.New("config does not exists")
)

type Repo interface {
	CreateSet(name string, ttl int) (domain.ConfigSet, error)
	GetSet(name string, maxAge int) (*domain.ConfigSet, error)
	GetSetNames(limit int, skip int) ([]string, error)
	DeleteSet(name string) (domain.ConfigSet, error)
	AddItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	UpdateItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	RemoveItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
}

type Secret interface {
	Get(name string) (string, error)
}

type Notifier interface {
	SetUpdated(name string, set domain.ConfigSet) error
}
