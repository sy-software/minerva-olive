package domain

import (
	"errors"
	"time"

	"github.com/sy-software/minerva-go-utils/datetime"
)

// ConfigType represents the posible ways to store config values
type ConfigType string

// Available ConfigTypes
const (
	// A config value stored in the secret manager
	Secret ConfigType = "secret"
	// A config value stored as plain value
	Plain ConfigType = "plain"
	// A config value containing a nested config set
	Nested ConfigType = "nested"
)

// Possible errors during config manipulation
var (
	// The config set already contains the given config key
	ErrDuplicatedKey = errors.New("duplicated config item key")
	ErrKeyNotExists  = errors.New("config item key does not exists")
	// A config item of type "nested" does not contain a string as value
	ErrInvalidNestedKeyValue = errors.New("invalid key value for nested config")
	// A config item of type "secret" does not contain a string as value
	ErrSecretKeyValue = errors.New("invalid key value for secret")
)

// ConfigItem represents a single config value
type ConfigItem struct {
	// The key to access this value inside a set
	Key string `json:"key"`
	// The stored value, for types "nested" and "secret" this should always be a string
	Value interface{} `json:"value"`
	// The type of config stored in this item
	Type ConfigType `json:"type"`
}

func NewConfigItem(key string, value interface{}, cfgType ConfigType) *ConfigItem {
	return &ConfigItem{
		Key:   key,
		Value: value,
		Type:  cfgType,
	}
}

type ConfigItemMap map[string]ConfigItem

// ConfigSet is a group of config values
type ConfigSet struct {
	// The name to address this set
	Name string `json:"name"`
	// When was this set created
	CreateDate time.Time `json:"createDate"`
	// When was this set last updated
	UpdateDate time.Time `json:"updateDate"`
	// The items contained in this set
	Items ConfigItemMap `json:"items"`
}

// NewConfigSet creates a new config set with the given items
func NewConfigSet(name string, items ...ConfigItem) *ConfigSet {
	mapItems := ConfigItemMap{}

	for _, i := range items {
		mapItems[i.Key] = i
	}

	return &ConfigSet{
		Name:       name,
		CreateDate: datetime.UnixUTCNow(),
		UpdateDate: datetime.UnixUTCNow(),
		Items:      mapItems,
	}
}

// Add saves the given item into this set
func (set *ConfigSet) Add(item ConfigItem) error {
	_, exists := set.Items[item.Key]

	if exists {
		return ErrDuplicatedKey
	}

	set.Items[item.Key] = item
	return nil
}

// Get finds a config item with the given key
// returns ErrKeyNotExists if the key is not present in this set
func (set *ConfigSet) Get(key string) (ConfigItem, error) {
	val, ok := set.Items[key]

	if !ok {
		return ConfigItem{}, ErrKeyNotExists
	}

	return val, nil
}

// Update replaces the config item with the same key as the one provided
// returns ErrKeyNotExists if the key is not present in this set
func (set *ConfigSet) Update(item ConfigItem) (ConfigItem, error) {
	_, ok := set.Items[item.Key]

	if !ok {
		return ConfigItem{}, ErrKeyNotExists
	}

	set.Items[item.Key] = item
	return item, nil
}

// Delete removes the config item with the given key
// returns ErrKeyNotExists if the key is not present in this set
func (set *ConfigSet) Delete(key string) (ConfigItem, error) {
	val, ok := set.Items[key]

	if !ok {
		return ConfigItem{}, ErrKeyNotExists
	}

	delete(set.Items, key)
	return val, nil
}
