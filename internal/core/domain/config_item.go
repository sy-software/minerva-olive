package domain

import "time"

type ConfigType string

const (
	Secret ConfigType = "secret"
	Plain  ConfigType = "plain"
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
