package service

import (
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
)

type ConfigService struct {
	repo          ports.Repo
	cache         ports.Repo
	secretManager ports.Secret
}

func NewConfigService(repo ports.Repo, cache ports.Repo, secretManager ports.Secret) *ConfigService {
	return &ConfigService{
		repo:          repo,
		cache:         cache,
		secretManager: secretManager,
	}
}

func (service *ConfigService) CreateSet(name string) (domain.ConfigSet, error) {
	return service.repo.CreateSet(name, ports.InfiniteTTL)
}

func (service *ConfigService) GetSet(name string) (domain.ConfigSet, error) {
	panic("not implemented")
}

func (service *ConfigService) GetSetNames(count int, skip int) ([]string, error) {
	panic("not implemented")
}

func (service *ConfigService) RenameSet(name string, newName string) (domain.ConfigSet, error) {
	panic("not implemented")
}

func (service *ConfigService) DeleteSet(name string) (domain.ConfigSet, error) {
	panic("not implemented")
}

func (service *ConfigService) AddItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
	panic("not implemented")
}

func (service *ConfigService) UpdateItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
	panic("not implemented")
}

func (service *ConfigService) RemoveItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
	panic("not implemented")
}

func (service *ConfigService) SetToJson(set domain.ConfigSet) ([]byte, error) {
	panic("not implemented")
}
