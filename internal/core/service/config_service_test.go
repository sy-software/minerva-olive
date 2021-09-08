package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sy-software/minerva-go-utils/datetime"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
	"github.com/sy-software/minerva-olive/mocks"
)

func TestCreateSet(t *testing.T) {
	config := domain.DefaultConfig()
	t.Run("Test a new config set can be created", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		now := datetime.UnixUTCNow()
		got, err := service.CreateSet(name)

		if err != nil {
			t.Errorf("Expected set to be created without errors, got: %v", err)
		}

		if got.Name != name {
			t.Errorf("Expected name: %q, got %q", name, got.Name)
		}

		if got.CreateDate.Before(now) {
			t.Errorf("Expected create date to be >= than: %q, got: %q", now, got.CreateDate)
		}

		if !got.UpdateDate.Equal(got.CreateDate) {
			t.Errorf("Expected update date (%q) to be equals to create date (%q)", got.UpdateDate, got.CreateDate)
		}
	})

	t.Run("Test a duplicated config can't be created", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()

		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)
		_, err := service.CreateSet(name)

		if err != ports.ErrDuplicatedConfig {
			t.Errorf("Expected error: %v got nil", ports.ErrDuplicatedConfig)
		}
	})
}

func TestReadSet(t *testing.T) {
	config := domain.DefaultConfig()
	t.Run("Test a set can be read", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		got, err := service.GetSet(name)
		if err != nil {
			t.Errorf("Expected set to be read without errors, got: %v", err)
		}

		if got.Name != name {
			t.Errorf("Expected name: %q, got %q", name, got.Name)
		}
	})

	t.Run("Test a non existing set can't be read", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		_, err := service.GetSet("otherName")

		if err == nil {
			t.Errorf("Expected error: %v got nil", ports.ErrConfigNotExists)
		}
	})

	t.Run("Test set names can be read and paginated", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		for i := 0; i < 20; i++ {
			service.CreateSet(fmt.Sprintf("mySet%d", i))
		}

		skip := 0
		count := 2
		names, err := service.GetSetNames(count, skip)

		if err != nil {
			t.Errorf("Expected set names to be read without errors, got: %v", err)
		}

		if len(names) != count {
			t.Errorf("Expected names length of: %d, got: %d", count, len(names))
		}

		expected := []string{"mySet0", "mySet1"}
		if !cmp.Equal(names, expected) {
			t.Errorf("Expected names: %+v, got: %+v", names, expected)
		}

		skip = 20
		count = 2
		names, err = service.GetSetNames(count, skip)

		if err != nil {
			t.Errorf("Expected set names to be read without errors, got: %v", err)
		}

		if len(names) != 0 {
			t.Errorf("Expected names length of: 0, got: %d", len(names))
		}
	})

	t.Run("Test a set can be read as json", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		items := []domain.ConfigItem{
			{
				Key:   "integer",
				Value: 100,
				Type:  domain.Plain,
			},
			{
				Key:   "string",
				Value: "Hello world!",
				Type:  domain.Plain,
			},
			{
				Key:   "bool",
				Value: true,
				Type:  domain.Plain,
			},
			{
				Key:   "float",
				Value: 42.5,
				Type:  domain.Plain,
			},
			{
				Key:   "array",
				Value: []int{1, 2, 4},
				Type:  domain.Plain,
			},
		}

		jsonMap := map[string]interface{}{}
		for _, item := range items {
			service.AddItem(item, name)
			jsonMap[item.Key] = item.Value
		}

		jsonBytes, _ := json.Marshal(jsonMap)

		got, err := service.GetSetJson(name, domain.AnyAge)
		if err != nil {
			t.Errorf("Expected set to be read without errors, got: %v", err)
		}

		if !cmp.Equal(got, jsonBytes) {
			t.Errorf("Expected json: %s, got %q", string(jsonBytes), got)
		}
	})
}

func TestDeleteSet(t *testing.T) {
	config := domain.DefaultConfig()
	t.Run("Test a set can be delete", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		_, err := service.DeleteSet(name)
		if err != nil {
			t.Errorf("Expected set to be delete without errors, got: %v", err)
		}

		_, err = service.GetSet(name)

		if err != ports.ErrConfigNotExists {
			t.Errorf("Expected error: %v got: %v", ports.ErrConfigNotExists, err)
		}
	})

	t.Run("Test a non existing set can't be deleted", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		_, err := service.DeleteSet("otherName")

		if err == nil {
			t.Errorf("Expected error: %v got nil", ports.ErrConfigNotExists)
		}
	})
}

func TestRenameSet(t *testing.T) {
	config := domain.DefaultConfig()
	t.Run("Test a set can be renamed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)
		newName := "newName"
		got, err := service.RenameSet(name, newName)
		if err != nil {
			t.Errorf("Expected set to be renamed without errors, got: %v", err)
		}

		if got.Name != newName {
			t.Errorf("Expected name: %q, got %q", newName, got.Name)
		}

		if !got.UpdateDate.After(got.CreateDate) {
			t.Errorf("Expected update date to be > %v, got: %v", got.CreateDate, got.UpdateDate)
		}

		got, err = service.GetSet(newName)
		if err != nil {
			t.Errorf("Expected set to be read without errors, got: %v", err)
		}

		if got.Name != newName {
			t.Errorf("Expected name: %q, got %q", newName, got.Name)
		}

		if !got.UpdateDate.After(got.CreateDate) {
			t.Errorf("Expected update date to be > %v, got: %v", got.CreateDate, got.UpdateDate)
		}

		_, err = service.GetSet(name)

		if err == nil {
			t.Errorf("Expected error: %v got nil", ports.ErrConfigNotExists)
		}
	})

	t.Run("Test a non existing set can't be renamed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)
		newName := "newName"
		_, err := service.RenameSet("notExists", newName)

		if err == nil {
			t.Errorf("Expected error: %v got nil", ports.ErrConfigNotExists)
		}
	})

	t.Run("Test a set can't be renamed to existing name", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)
		newName := "newName"
		service.CreateSet(newName)
		_, err := service.RenameSet("notExists", newName)

		if err == nil {
			t.Errorf("Expected error: %v got nil", ports.ErrDuplicatedConfig)
		}
	})
}

/// Test items

func TestAddItemToSet(t *testing.T) {
	config := domain.DefaultConfig()
	t.Run("Test a new key can be added", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)
		key := "myKey"
		newItem := domain.ConfigItem{
			Key:   key,
			Value: 100,
			Type:  domain.Plain,
		}
		got, err := service.AddItem(newItem, name)
		if err != nil {
			t.Errorf("Expected item to be added without errors, got: %v", err)
		}

		if got.Name != name {
			t.Errorf("Expected name: %q, got %q", name, got.Name)
		}

		if !got.UpdateDate.After(got.CreateDate) {
			t.Errorf("Expected update date to be > %v, got: %v", got.CreateDate, got.UpdateDate)
		}

		gotItem, err := got.Get(key)
		if err != nil {
			t.Errorf("Expected item can be getted without errors, got: %v", err)
		}

		if !cmp.Equal(gotItem, newItem) {
			t.Errorf("Expected item: %+v, got: %+v", newItem, gotItem)
		}
	})

	t.Run("Test a duplicated key can't be added", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		key := "myKey"
		newItem := domain.ConfigItem{
			Key:   key,
			Value: 100,
			Type:  domain.Plain,
		}
		service.AddItem(newItem, name)
		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)
		now := datetime.UnixUTCNow()
		newItem.Value = 101
		got, err := service.AddItem(newItem, name)
		if err == nil {
			t.Errorf("Expected error: %v, got nil", domain.ErrDuplicatedKey)
		}

		if got.Name != name {
			t.Errorf("Expected name: %q, got %q", name, got.Name)
		}

		if !got.UpdateDate.Before(now) {
			t.Errorf("Expected update date to be %v, got: %v", got.CreateDate, got.UpdateDate)
		}

		gotItem, err := got.Get(key)
		if err != nil {
			t.Errorf("Expected item can be getted without errors, got: %v", err)
		}

		if gotItem.Value != 100 {
			t.Errorf("Expected item value: %+v, got: %+v", 100, gotItem.Value)
		}
	})
}

func TestUpdateItemFromSet(t *testing.T) {
	config := domain.DefaultConfig()
	t.Run("Test a key can be updated", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		key := "myKey"
		newItem := domain.ConfigItem{
			Key:   key,
			Value: 100,
			Type:  domain.Plain,
		}
		service.AddItem(newItem, name)

		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)

		newItem.Value = 1000
		got, err := service.UpdateItem(newItem, name)
		if err != nil {
			t.Errorf("Expected item to be updated without errors, got: %v", err)
		}

		if got.Name != name {
			t.Errorf("Expected name: %q, got %q", name, got.Name)
		}

		if !got.UpdateDate.After(got.CreateDate) {
			t.Errorf("Expected update date to be > %v, got: %v", got.CreateDate, got.UpdateDate)
		}

		gotItem, err := got.Get(key)
		if err != nil {
			t.Errorf("Expected item can be getted without errors, got: %v", err)
		}

		if !cmp.Equal(gotItem, newItem) {
			t.Errorf("Expected item: %+v, got: %+v", newItem, gotItem)
		}
	})

	t.Run("Test a non existing key can't be updated", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		key := "myKey"
		newItem := domain.ConfigItem{
			Key:   key,
			Value: 100,
			Type:  domain.Plain,
		}
		service.AddItem(newItem, name)
		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)
		now := datetime.UnixUTCNow()
		newItem.Key = "otherKey"
		newItem.Value = 101
		got, err := service.UpdateItem(newItem, name)
		if err == nil {
			t.Errorf("Expected error: %v, got nil", domain.ErrKeyNotExists)
		}

		if got.Name != name {
			t.Errorf("Expected name: %q, got %q", name, got.Name)
		}

		if !got.UpdateDate.Before(now) {
			t.Errorf("Expected update date to be %v, got: %v", got.CreateDate, got.UpdateDate)
		}

		gotItem, err := got.Get(key)
		if err != nil {
			t.Errorf("Expected item can be getted without errors, got: %v", err)
		}

		if gotItem.Value != 100 {
			t.Errorf("Expected item value: %+v, got: %+v", 100, gotItem.Value)
		}
	})
}

func TestRemoveItemFromSet(t *testing.T) {
	config := domain.DefaultConfig()
	t.Run("Test a key can be removed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		key := "myKey"
		newItem := domain.ConfigItem{
			Key:   key,
			Value: 100,
			Type:  domain.Plain,
		}
		service.AddItem(newItem, name)

		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)

		newItem.Value = 1000
		got, err := service.RemoveItem(newItem, name)
		if err != nil {
			t.Errorf("Expected item to be removed without errors, got: %v", err)
		}

		if got.Name != name {
			t.Errorf("Expected name: %q, got %q", name, got.Name)
		}

		if !got.UpdateDate.After(got.CreateDate) {
			t.Errorf("Expected update date to be > %v, got: %v", got.CreateDate, got.UpdateDate)
		}

		_, err = got.Get(key)
		if err != domain.ErrKeyNotExists {
			t.Errorf("Expected error: %v, got: %v", domain.ErrKeyNotExists, err)
		}

	})

	t.Run("Test a non existing key can't be removed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		key := "myKey"
		newItem := domain.ConfigItem{
			Key:   key,
			Value: 100,
			Type:  domain.Plain,
		}
		service.AddItem(newItem, name)
		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)
		now := datetime.UnixUTCNow()
		newItem.Key = "otherKey"
		newItem.Value = 101
		got, err := service.RemoveItem(newItem, name)
		if err == nil {
			t.Errorf("Expected error: %v, got nil", domain.ErrKeyNotExists)
		}

		if got.Name != name {
			t.Errorf("Expected name: %q, got %q", name, got.Name)
		}

		if !got.UpdateDate.Before(now) {
			t.Errorf("Expected update date to be %v, got: %v", got.CreateDate, got.UpdateDate)
		}

		gotItem, err := got.Get(key)
		if err != nil {
			t.Errorf("Expected item can be getted without errors, got: %v", err)
		}

		if gotItem.Value != 100 {
			t.Errorf("Expected item value: %+v, got: %+v", 100, gotItem.Value)
		}
	})
}

/// Test JSON serialization

func TestConvertSetToJSON(t *testing.T) {
	config := domain.DefaultConfig()
	t.Run("Test single level config set is converted", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		items := []domain.ConfigItem{
			{
				Key:   "integer",
				Value: 100,
				Type:  domain.Plain,
			},
			{
				Key:   "string",
				Value: "Hello world!",
				Type:  domain.Plain,
			},
			{
				Key:   "bool",
				Value: true,
				Type:  domain.Plain,
			},
			{
				Key:   "float",
				Value: 42.5,
				Type:  domain.Plain,
			},
			{
				Key:   "array",
				Value: []int{1, 2, 4},
				Type:  domain.Plain,
			},
		}

		jsonMap := map[string]interface{}{}
		for _, item := range items {
			service.AddItem(item, name)
			jsonMap[item.Key] = item.Value
		}

		jsonBytes, _ := json.Marshal(jsonMap)

		set, _ := service.GetSet(name)
		got, err := service.SetToJson(set)
		if err != nil {
			t.Errorf("Expected set to be serialized without errors, got: %v", err)
		}

		if !cmp.Equal(jsonBytes, got) {
			t.Errorf("Expected json: %s, got: %v", string(jsonBytes), string(got))
		}
	})

	t.Run("Test multi level config set is converted", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		nested := "nested"
		service.CreateSet(name)
		service.CreateSet(nested)

		items := []domain.ConfigItem{
			{
				Key:   "integer",
				Value: 100,
				Type:  domain.Plain,
			},
			{
				Key:   "string",
				Value: "Hello world!",
				Type:  domain.Plain,
			},
			{
				Key:   "bool",
				Value: true,
				Type:  domain.Plain,
			},
			{
				Key:   "float",
				Value: 42.5,
				Type:  domain.Plain,
			},
			{
				Key:   "array",
				Value: []int{1, 2, 4},
				Type:  domain.Plain,
			},
		}

		nestedMap := map[string]interface{}{}
		for _, item := range items {
			service.AddItem(item, nested)
			nestedMap[item.Key] = item.Value
		}

		jsonMap := map[string]interface{}{}
		service.AddItem(domain.ConfigItem{
			Key:   nested,
			Value: nested,
			Type:  domain.Nested,
		}, name)
		jsonMap[nested] = nestedMap

		jsonBytes, _ := json.Marshal(jsonMap)

		set, _ := service.GetSet(name)
		got, err := service.SetToJson(set)
		if err != nil {
			t.Errorf("Expected set to be serialized without errors, got: %v", err)
		}

		if !cmp.Equal(jsonBytes, got) {
			t.Errorf("Expected json: %s, got: %v", string(jsonBytes), string(got))
		}
	})

	t.Run("Test secret are expanded using secret manager", func(t *testing.T) {
		values := map[string]string{
			"integer": "100",
			"string":  "Hello world!",
			"bool":    "true",
			"float":   "42.5",
		}

		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{
			Values: values,
		}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		items := []domain.ConfigItem{
			{
				Key:   "integer",
				Value: "integer",
				Type:  domain.Secret,
			},
			{
				Key:   "string",
				Value: "string",
				Type:  domain.Secret,
			},
			{
				Key:   "bool",
				Value: "bool",
				Type:  domain.Secret,
			},
			{
				Key:   "float",
				Value: "float",
				Type:  domain.Secret,
			},
		}

		jsonMap := map[string]interface{}{}
		for _, item := range items {
			service.AddItem(item, name)
			jsonMap[item.Key] = values[item.Key]
		}

		jsonBytes, _ := json.Marshal(jsonMap)

		set, _ := service.GetSet(name)
		got, err := service.SetToJson(set)
		if err != nil {
			t.Errorf("Expected set to be serialized without errors, got: %v", err)
		}

		if !cmp.Equal(jsonBytes, got) {
			t.Errorf("Expected json: %s, got: %v", string(jsonBytes), string(got))
		}
	})
}

/// Test Caching of values

func TestJSONIsSavedToCache(t *testing.T) {
	t.Run("Test a new config is saved to cache after creation", func(t *testing.T) {
		config := domain.DefaultConfig()
		config.CacheTTL = time.Duration(10) * time.Second
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()

		name := "mySet"

		cacheSaveCalled := false
		cacheRepo.SaveJSONInterceptor = func(json []byte, key string, ttl int) error {
			cacheSaveCalled = true

			if string(json) != "{}" {
				t.Errorf("Expected json: %s, got: %s", "{}", string(json))
			}

			if ttl != int(config.CacheTTL) {
				t.Errorf("Expected TTL: %d, got: %d", config.CacheTTL, ttl)
			}

			if !strings.Contains(key, name) {
				t.Errorf("Expected cache key to contain: %q, got: %q", name, key)
			}

			return nil
		}
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		_, err := service.CreateSet(name)

		if err != nil {
			t.Errorf("Expected set to be created without errors, got: %v", err)
		}

		if !cacheSaveCalled {
			t.Errorf("Expected cache SaveJSON to be called")
		}
	})

	t.Run("Test renamed set is updated in cache", func(t *testing.T) {
		config := domain.DefaultConfig()
		config.CacheTTL = time.Duration(10) * time.Second
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}

		name := "mySet"
		newName := "newName"
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)
		service.CreateSet(name)

		cacheSaveCalled := false
		cacheRepo.SaveJSONInterceptor = func(json []byte, key string, ttl int) error {
			cacheSaveCalled = true

			if string(json) != "{}" {
				t.Errorf("Expected json: %s, got: %s", "{}", string(json))
			}

			if ttl != int(config.CacheTTL) {
				t.Errorf("Expected TTL: %d, got: %d", config.CacheTTL, ttl)
			}

			if !strings.Contains(key, newName) {
				t.Errorf("Expected cache key to contain: %q, got: %q", newName, key)
			}

			return nil
		}
		cacheRemoveCalled := false
		cacheRepo.RemoveJSONInterceptor = func(key string) error {
			if !strings.Contains(key, name) {
				t.Errorf("Expected cache key to contain: %q, got: %q", name, key)
			}
			cacheRemoveCalled = true
			return nil
		}

		_, err := service.RenameSet(name, newName)
		if err != nil {
			t.Errorf("Expected set to be renamed without errors, got: %v", err)
		}

		if !cacheSaveCalled {
			t.Errorf("Expected cache SaveJSON to be called")
		}

		if !cacheRemoveCalled {
			t.Errorf("Expected cache RemoveJSON to be called")
		}
	})

	t.Run("Test adding item to set updates cache", func(t *testing.T) {
		config := domain.DefaultConfig()
		config.CacheTTL = time.Duration(10) * time.Second
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}

		name := "mySet"
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)
		service.CreateSet(name)

		cacheSaveCalled := false
		cacheRepo.SaveJSONInterceptor = func(json []byte, key string, ttl int) error {
			cacheSaveCalled = true

			if string(json) != `{"string":"hello"}` {
				t.Errorf("Expected json: %s, got: %s", "{}", string(json))
			}

			if ttl != int(config.CacheTTL) {
				t.Errorf("Expected TTL: %d, got: %d", config.CacheTTL, ttl)
			}

			if !strings.Contains(key, name) {
				t.Errorf("Expected cache key to contain: %q, got: %q", name, key)
			}

			return nil
		}

		_, err := service.AddItem(domain.ConfigItem{
			Key:   "string",
			Value: "hello",
			Type:  domain.Plain,
		}, name)
		if err != nil {
			t.Errorf("Expected set to be updated without errors, got: %v", err)
		}

		if !cacheSaveCalled {
			t.Errorf("Expected cache SaveJSON to be called")
		}
	})

	t.Run("Test updating item in set updates cache", func(t *testing.T) {
		config := domain.DefaultConfig()
		config.CacheTTL = time.Duration(10) * time.Second
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}

		name := "mySet"
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)
		service.CreateSet(name)
		service.AddItem(domain.ConfigItem{
			Key:   "string",
			Value: "hello",
			Type:  domain.Plain,
		}, name)

		cacheSaveCalled := false
		cacheRepo.SaveJSONInterceptor = func(json []byte, key string, ttl int) error {
			cacheSaveCalled = true

			if string(json) != `{"string":"goodbye"}` {
				t.Errorf("Expected json: %s, got: %s", "{}", string(json))
			}

			if ttl != int(config.CacheTTL) {
				t.Errorf("Expected TTL: %d, got: %d", config.CacheTTL, ttl)
			}

			if !strings.Contains(key, name) {
				t.Errorf("Expected cache key to contain: %q, got: %q", name, key)
			}

			return nil
		}

		_, err := service.UpdateItem(domain.ConfigItem{
			Key:   "string",
			Value: "goodbye",
			Type:  domain.Plain,
		}, name)
		if err != nil {
			t.Errorf("Expected set to be updated without errors, got: %v", err)
		}

		if !cacheSaveCalled {
			t.Errorf("Expected cache SaveJSON to be called")
		}
	})

	t.Run("Test removing item from set updates cache", func(t *testing.T) {
		config := domain.DefaultConfig()
		config.CacheTTL = time.Duration(10) * time.Second
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}

		name := "mySet"
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)
		service.CreateSet(name)
		service.AddItem(domain.ConfigItem{
			Key:   "string",
			Value: "hello",
			Type:  domain.Plain,
		}, name)

		cacheSaveCalled := false
		cacheRepo.SaveJSONInterceptor = func(json []byte, key string, ttl int) error {
			cacheSaveCalled = true

			if string(json) != "{}" {
				t.Errorf("Expected json: %s, got: %s", "{}", string(json))
			}

			if ttl != int(config.CacheTTL) {
				t.Errorf("Expected TTL: %d, got: %d", config.CacheTTL, ttl)
			}

			if !strings.Contains(key, name) {
				t.Errorf("Expected cache key to contain: %q, got: %q", name, key)
			}

			return nil
		}

		_, err := service.RemoveItem(domain.ConfigItem{
			Key:   "string",
			Value: "hello",
			Type:  domain.Plain,
		}, name)
		if err != nil {
			t.Errorf("Expected set to be updated without errors, got: %v", err)
		}

		if !cacheSaveCalled {
			t.Errorf("Expected cache SaveJSON to be called")
		}
	})
}

func TestJSONIsReadFromCache(t *testing.T) {
	t.Run("Test JSON is returned from cache", func(t *testing.T) {
		config := domain.DefaultConfig()
		config.CacheTTL = time.Duration(10) * time.Second
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()

		name := "mySet"
		expectedMaxAge := domain.AnyAge
		cacheGetCalled := false
		cacheRepo.GetJSONInterceptor = func(key string, maxAge int) ([]byte, error) {
			cacheGetCalled = true
			if !strings.Contains(key, name) {
				t.Errorf("Expected cache key to contain: %q, got: %q", name, key)
			}

			if maxAge != expectedMaxAge {
				t.Errorf("Expected MaxAge: %d, got: %d", expectedMaxAge, maxAge)
			}

			return []byte("{}"), nil
		}

		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)
		service.CreateSet(name)

		jsonBytes, err := service.GetSetJson(name, expectedMaxAge)
		if err != nil {
			t.Errorf("Expected set to be retrieved without errors, got: %v", err)
		}

		if !cacheGetCalled {
			t.Errorf("Expected cache GetJSON to be called")
		}

		if string(jsonBytes) != "{}" {
			t.Errorf("Expected json: %s, got: %s", "{}", string(jsonBytes))
		}
	})

	t.Run("Test JSON is returned from persintent storage if cache is not present", func(t *testing.T) {
		config := domain.DefaultConfig()
		config.CacheTTL = time.Duration(10) * time.Second
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()

		name := "mySet"
		expectedMaxAge := domain.AnyAge
		cacheGetCalled := false
		cacheRepo.GetJSONInterceptor = func(key string, maxAge int) ([]byte, error) {
			cacheGetCalled = true
			if !strings.Contains(key, name) {
				t.Errorf("Expected cache key to contain: %q, got: %q", name, key)
			}

			if maxAge != expectedMaxAge {
				t.Errorf("Expected MaxAge: %d, got: %d", expectedMaxAge, maxAge)
			}

			return []byte("don't return this"), ports.ErrConfigNotExists
		}

		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)
		service.CreateSet(name)

		jsonBytes, err := service.GetSetJson(name, expectedMaxAge)
		if err != nil {
			t.Errorf("Expected set to be retrieved without errors, got: %v", err)
		}

		if !cacheGetCalled {
			t.Errorf("Expected cache GetJSON to be called")
		}

		if string(jsonBytes) != "{}" {
			t.Errorf("Expected json: %s, got: %s", "{}", string(jsonBytes))
		}
	})

	t.Run("Test JSON is returned from persintent storage if cache is older than max age", func(t *testing.T) {
		config := domain.DefaultConfig()
		config.CacheTTL = time.Duration(1) * time.Second
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()

		name := "mySet"
		expectedMaxAge := time.Duration(1) * time.Second
		cacheGetCalled := false
		cacheRepo.GetJSONInterceptor = func(key string, maxAge int) ([]byte, error) {
			cacheGetCalled = true
			if !strings.Contains(key, name) {
				t.Errorf("Expected cache key to contain: %q, got: %q", name, key)
			}

			if time.Duration(maxAge) != expectedMaxAge {
				t.Errorf("Expected MaxAge: %d, got: %d", expectedMaxAge, maxAge)
			}

			return []byte("don't return this"), ports.ErrOldValue
		}

		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)
		service.CreateSet(name)

		time.Sleep(2 * time.Second)
		jsonBytes, err := service.GetSetJson(name, int(expectedMaxAge))
		if err != nil {
			t.Errorf("Expected set to be retrieved without errors, got: %v", err)
		}

		if !cacheGetCalled {
			t.Errorf("Expected cache GetJSON to be called")
		}

		if string(jsonBytes) != "{}" {
			t.Errorf("Expected json: %s, got: %s", "{}", string(jsonBytes))
		}
	})
}
