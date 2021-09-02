package domain

import (
	"errors"
	"time"

	"github.com/sy-software/minerva-go-utils/datetime"
)

type ConfigType string

const (
	Secret ConfigType = "secret"
	Plain  ConfigType = "plain"
	Nested ConfigType = "nested"
)

var (
	ErrDuplicatedKey         = errors.New("duplicated config item key")
	ErrKeyNotExists          = errors.New("config item key does not exists")
	ErrInvalidNestedKeyValue = errors.New("invalid key value for nested config")
	ErrSecretKeyValue        = errors.New("invalid key value for secret")
)

type ConfigItem struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Type  ConfigType  `json:"type"`
}

func NewConfigItem(key string, value interface{}, cfgType ConfigType) *ConfigItem {
	return &ConfigItem{
		Key:   key,
		Value: value,
		Type:  cfgType,
	}
}

type ConfigItemMap map[string]ConfigItem

type ConfigSet struct {
	Name       string
	CreateDate time.Time
	UpdateDate time.Time
	Items      ConfigItemMap
}

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

func (set *ConfigSet) Add(item ConfigItem) error {
	_, exists := set.Items[item.Key]

	if exists {
		return ErrDuplicatedKey
	}

	set.Items[item.Key] = item
	return nil
}

func (set *ConfigSet) Get(key string) (ConfigItem, error) {
	val, ok := set.Items[key]

	if !ok {
		return ConfigItem{}, ErrKeyNotExists
	}

	return val, nil
}

func (set *ConfigSet) Update(item ConfigItem) (ConfigItem, error) {
	_, ok := set.Items[item.Key]

	if !ok {
		return ConfigItem{}, ErrKeyNotExists
	}

	set.Items[item.Key] = item
	return item, nil
}

func (set *ConfigSet) Delete(key string) (ConfigItem, error) {
	val, ok := set.Items[key]

	if !ok {
		return ConfigItem{}, ErrKeyNotExists
	}

	delete(set.Items, key)
	return val, nil
}
