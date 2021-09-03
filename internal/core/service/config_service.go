package service

import (
	"encoding/json"

	"github.com/sy-software/minerva-go-utils/datetime"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
)

type ConfigService struct {
	repo          ports.Repo
	cache         ports.CacheRepo
	secretManager ports.Secret
}

func NewConfigService(repo ports.Repo, cache ports.CacheRepo, secretManager ports.Secret) *ConfigService {
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

	return service.repo.CreateSet(newSet)
}

func (service *ConfigService) GetSet(name string) (domain.ConfigSet, error) {
	set, err := service.repo.GetSet(name)
	if err == nil {
		return *set, err
	}

	return domain.ConfigSet{}, err
}

func (service *ConfigService) GetSetJson(name string) ([]byte, error) {
	set, err := service.GetSet(name)
	if err != nil {
		return []byte{}, err
	}

	return service.SetToJson(set)
}

func (service *ConfigService) GetSetNames(count int, skip int) ([]string, error) {
	return service.repo.GetSetNames(count, skip)
}

func (service *ConfigService) RenameSet(name string, newName string) (domain.ConfigSet, error) {
	set, err := service.repo.GetSet(name)
	if err != nil {
		return domain.ConfigSet{}, err
	}

	_, err = service.repo.GetSet(newName)
	if err != ports.ErrConfigNotExists {
		return domain.ConfigSet{}, ports.ErrDuplicatedConfig
	}

	_, err = service.DeleteSet(name)
	if err != nil {
		return domain.ConfigSet{}, err
	}

	set.Name = newName
	set.UpdateDate = datetime.UnixUTCNow()
	return service.repo.CreateSet(*set)
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
	mappedItems, err := service.setToMap(set)
	if err != nil {
		return []byte{}, err
	}

	return json.Marshal(mappedItems)
}

func (service *ConfigService) setToMap(set domain.ConfigSet) (map[string]interface{}, error) {
	mappedItems := map[string]interface{}{}
	for _, item := range set.Items {

		switch item.Type {
		case domain.Nested:
			name, ok := item.Value.(string)
			if !ok {
				return mappedItems, domain.ErrInvalidNestedKeyValue
			}
			set, err := service.GetSet(name)
			if err != nil {
				return mappedItems, err
			}

			mappedItems[item.Key], err = service.setToMap(set)
		case domain.Secret:
			name, ok := item.Value.(string)
			if !ok {
				return mappedItems, domain.ErrSecretKeyValue
			}
			val, err := service.secretManager.Get(name)

			if err != nil {
				return mappedItems, err
			}

			mappedItems[item.Key] = val
		default:
			mappedItems[item.Key] = item.Value
		}
	}

	return mappedItems, nil
}
