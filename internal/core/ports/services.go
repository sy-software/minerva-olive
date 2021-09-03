package ports

import "github.com/sy-software/minerva-olive/internal/core/domain"

type ConfigService interface {
	CreateSet(name string) (domain.ConfigSet, error)
	GetSet(name string) (domain.ConfigSet, error)
	GetSetJson(name string) ([]byte, error)
	GetSetNames(count int, skip int) ([]string, error)
	RenameSet(name string, newName string) (domain.ConfigSet, error)
	DeleteSet(name string) (domain.ConfigSet, error)
	AddItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	UpdateItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	RemoveItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	SetToJson(set domain.ConfigSet) ([]byte, error)
}
