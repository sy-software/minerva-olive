package service

import (
	"github.com/sy-software/minerva-go-utils/datetime"
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
	now := datetime.UnixUTCNow()
	newSet := domain.ConfigSet{
		Name:       name,
		Items:      domain.ConfigItemMap{},
		CreateDate: now,
		UpdateDate: now,
	}

	return service.repo.CreateSet(newSet, ports.InfiniteTTL)
}

func (service *ConfigService) GetSet(name string) (domain.ConfigSet, error) {
	set, err := service.repo.GetSet(name, ports.AnyAge)
	if err == nil {
		return *set, err
	}

	return domain.ConfigSet{}, err
}

func (service *ConfigService) GetSetNames(count int, skip int) ([]string, error) {
	return service.repo.GetSetNames(count, skip)
}

func (service *ConfigService) RenameSet(name string, newName string) (domain.ConfigSet, error) {
	set, err := service.repo.GetSet(name, ports.AnyAge)
	if err != nil {
		return domain.ConfigSet{}, err
	}

	_, err = service.repo.GetSet(newName, ports.AnyAge)
	if err != ports.ErrConfigNotExists {
		return domain.ConfigSet{}, ports.ErrDuplicatedConfig
	}

	_, err = service.DeleteSet(name)
	if err != nil {
		return domain.ConfigSet{}, err
	}

	set.Name = newName
	set.UpdateDate = datetime.UnixUTCNow()
	return service.repo.CreateSet(*set, ports.InfiniteTTL)
}

func (service *ConfigService) DeleteSet(name string) (domain.ConfigSet, error) {
	return service.repo.DeleteSet(name)
}

func (service *ConfigService) AddItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
	set, err := service.repo.AddItem(item, setName)
	if err == domain.ErrDuplicatedKey {
		set, _ = service.GetSet(setName)
		return set, err
	}

	return set, err
}

func (service *ConfigService) UpdateItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
	set, err := service.repo.UpdateItem(item, setName)
	if err == domain.ErrKeyNotExists {
		set, _ = service.GetSet(setName)
		return set, err
	}

	return set, err
}

func (service *ConfigService) RemoveItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
	set, err := service.repo.RemoveItem(item, setName)
	if err == domain.ErrKeyNotExists {
		set, _ = service.GetSet(setName)
		return set, err
	}

	return set, err
}

func (service *ConfigService) SetToJson(set domain.ConfigSet) ([]byte, error) {
	panic("not implemented")
}
