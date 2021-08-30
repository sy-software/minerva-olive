package mocks

import (
	"sort"

	"github.com/sy-software/minerva-go-utils/datetime"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
)

type MemRepo struct {
	Sets map[string]*domain.ConfigSet

	CreateSetInterceptor   func(set domain.ConfigSet, ttl int) (domain.ConfigSet, error)
	GetSetInterceptor      func(name string, maxAge int) (*domain.ConfigSet, error)
	GetSetNamesInterceptor func(count int, skip int) ([]string, error)
	DeleteSetInterceptor   func(name string) (domain.ConfigSet, error)
	AddItemInterceptor     func(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	UpdateItemInterceptor  func(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	RemoveItemInterceptor  func(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
}

func NewMockRepo() *MemRepo {
	return &MemRepo{
		Sets: make(map[string]*domain.ConfigSet),
	}
}

func (repo *MemRepo) CreateSet(set domain.ConfigSet, ttl int) (domain.ConfigSet, error) {
	if repo.CreateSetInterceptor != nil {
		return repo.CreateSetInterceptor(set, ttl)
	}

	_, exists := repo.Sets[set.Name]

	if exists {
		return domain.ConfigSet{}, ports.ErrDuplicatedConfig
	}

	repo.Sets[set.Name] = &set
	return set, nil
}

func (repo *MemRepo) GetSet(name string, maxAge int) (*domain.ConfigSet, error) {
	if repo.GetSetInterceptor != nil {
		return repo.GetSetInterceptor(name, maxAge)
	}

	value, exists := repo.Sets[name]

	if !exists {
		return nil, ports.ErrConfigNotExists
	}

	return value, nil
}

func (repo *MemRepo) GetSetNames(limit int, skip int) ([]string, error) {
	if repo.GetSetNamesInterceptor != nil {
		return repo.GetSetNamesInterceptor(limit, skip)
	}

	var keys []string
	for k := range repo.Sets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if skip >= len(keys) {
		return []string{}, nil
	}

	available := len(keys) - skip
	capLimit := limit
	if limit > available {
		capLimit = available
	}

	return keys[skip : skip+capLimit], nil
}

func (repo *MemRepo) DeleteSet(name string) (domain.ConfigSet, error) {
	if repo.DeleteSetInterceptor != nil {
		return repo.DeleteSetInterceptor(name)
	}

	value, ok := repo.Sets[name]
	if ok {
		delete(repo.Sets, name)
		return *value, nil
	}

	return domain.ConfigSet{}, ports.ErrConfigNotExists
}

func (repo *MemRepo) AddItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
	if repo.AddItemInterceptor != nil {
		return repo.AddItemInterceptor(item, setName)
	}

	set, err := repo.GetSet(setName, ports.AnyAge)

	if err != nil {
		return domain.ConfigSet{}, err
	}

	err = set.Add(item)
	if err != nil {
		return domain.ConfigSet{}, err
	}

	set.UpdateDate = datetime.UnixUTCNow()
	return *set, nil
}

func (repo *MemRepo) UpdateItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
	if repo.UpdateItemInterceptor != nil {
		return repo.UpdateItemInterceptor(item, setName)
	}

	set, err := repo.GetSet(setName, ports.AnyAge)

	if err != nil {
		return domain.ConfigSet{}, err
	}

	_, err = set.Update(item)
	if err != nil {
		return domain.ConfigSet{}, err
	}

	set.UpdateDate = datetime.UnixUTCNow()
	return *set, nil
}

func (repo *MemRepo) RemoveItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error) {
	if repo.RemoveItemInterceptor != nil {
		return repo.RemoveItemInterceptor(item, setName)
	}

	set, err := repo.GetSet(setName, ports.AnyAge)

	if err != nil {
		return domain.ConfigSet{}, err
	}

	_, err = set.Delete(item.Key)
	if err != nil {
		return domain.ConfigSet{}, err
	}

	set.UpdateDate = datetime.UnixUTCNow()
	return *set, nil
}
