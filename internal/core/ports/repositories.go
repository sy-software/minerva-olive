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

// Repo is an interface to apply CRUD operations over ConfigSet and ConfigItem
type Repo interface {
	// CreateSet Saves a new ConfigSet
	CreateSet(set domain.ConfigSet) (domain.ConfigSet, error)
	// GetSet Finds a ConfigSet by name
	GetSet(name string) (*domain.ConfigSet, error)
	// GetSetNames returns all stored ConfigSet names paginated
	GetSetNames(limit int, skip int) ([]string, error)
	// DeleteSet removes the ConfigSet with the given name
	DeleteSet(name string) (domain.ConfigSet, error)
	// AddItem inserts the given ConfigItem into the ConfigSet with setName
	AddItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	// UpdateItem updates the given ConfigItem into the ConfigSet with setName
	UpdateItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	// RemoveItem removes the given ConfigItem from the ConfigSet with setName
	RemoveItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
}

// CacheRepo can be implemented to provide manage JSON objects cache
type CacheRepo interface {
	// SaveJSON stores json bytes into a cache key with a given ttl
	// If the ttl is equals to -1 the key should be stored without expiration
	SaveJSON(json []byte, key string, ttl int) error
	// GetJSON retrives the json bytes from the provided cache key
	GetJSON(key string, maxAge int) ([]byte, error)
	// RemoveJSON deletes the value of the given cache key
	RemoveJSON(key string) error
}

// Secret is used to retrive the real value of ConfigItem of type secret
type Secret interface {
	Get(name string) (string, error)
}

// ToggleRepo provide operations to manage feature flags
type ToggleRepo interface {
	// GetFlag with the given name, using the context to provide information for rule evaluations
	GetFlag(name string, ctx context.Context) domain.ToggleFlag
	// GetFlagWithDefaults with the given name. If the flag is not present the provided defaults are returned
	GetFlagWithDefaults(name string, defStatus bool, defData interface{}, ctx context.Context) domain.ToggleFlag
	// SetFlag saves a flag with the provided information
	SetFlag(name string, status bool, data interface{}) error
}

type Notifier interface {
	SetUpdated(name string, set domain.ConfigSet) error
}
