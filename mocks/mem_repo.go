package mocks

import (
	"sort"
	"time"

	"github.com/sy-software/minerva-go-utils/datetime"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
)

type CacheItem struct {
	val      []byte
	creation time.Time
	ttl      time.Duration
}

type MemRepo struct {
	Sets  map[string]*domain.ConfigSet
	Cache map[string]CacheItem

	CreateSetInterceptor   func(set domain.ConfigSet) (domain.ConfigSet, error)
	GetSetInterceptor      func(name string) (*domain.ConfigSet, error)
	GetSetNamesInterceptor func(count int, skip int) ([]string, error)
	DeleteSetInterceptor   func(name string) (domain.ConfigSet, error)
	AddItemInterceptor     func(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	UpdateItemInterceptor  func(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	RemoveItemInterceptor  func(item domain.ConfigItem, setName string) (domain.ConfigSet, error)

	SaveJSONInterceptor   func(json []byte, key string, ttl int) error
	GetJSONInterceptor    func(key string, maxAge int) ([]byte, error)
	RemoveJSONInterceptor func(key string) error
}

func NewMockRepo() *MemRepo {
	return &MemRepo{
		Sets:  make(map[string]*domain.ConfigSet),
		Cache: make(map[string]CacheItem),
	}
}

func (repo *MemRepo) CreateSet(set domain.ConfigSet) (domain.ConfigSet, error) {
	if repo.CreateSetInterceptor != nil {
		return repo.CreateSetInterceptor(set)
	}

	_, exists := repo.Sets[set.Name]

	if exists {
		return domain.ConfigSet{}, ports.ErrDuplicatedConfig
	}

	repo.Sets[set.Name] = &set
	return set, nil
}

func (repo *MemRepo) GetSet(name string) (*domain.ConfigSet, error) {
	if repo.GetSetInterceptor != nil {
		return repo.GetSetInterceptor(name)
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

	set, err := repo.GetSet(setName)

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

	set, err := repo.GetSet(setName)

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

	set, err := repo.GetSet(setName)

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

// Cache

func (repo *MemRepo) SaveJSON(json []byte, key string, ttl int) error {
	if repo.SaveJSONInterceptor != nil {
		return repo.SaveJSONInterceptor(json, key, ttl)
	}

	repo.Cache[key] = CacheItem{
		val:      json,
		creation: datetime.UnixUTCNow(),
		ttl:      time.Duration(ttl),
	}

	return nil
}

func (repo *MemRepo) GetJSON(key string, maxAge int) ([]byte, error) {
	if repo.GetJSONInterceptor != nil {
		return repo.GetJSONInterceptor(key, maxAge)
	}

	value, exists := repo.Cache[key]

	if !exists {
		return nil, ports.ErrConfigNotExists
	}

	if maxAge != domain.AnyAge {
		now := datetime.UnixUTCNow()
		if value.creation.Add(time.Duration(maxAge)).Before(now) {
			return nil, ports.ErrOldValue
		}
	}

	return value.val, nil
}

func (repo *MemRepo) RemoveJSON(key string) error {
	if repo.RemoveJSONInterceptor != nil {
		return repo.RemoveJSONInterceptor(key)
	}

	delete(repo.Cache, key)

	return nil
}
