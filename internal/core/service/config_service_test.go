package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sy-software/minerva-go-utils/datetime"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
	"github.com/sy-software/minerva-olive/mocks"
)

func TestCreateSet(t *testing.T) {
	t.Run("Test a new config set can be created", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)
		_, err := service.CreateSet(name)

		if err == nil {
			t.Errorf("Expected error: %v got nil", ports.ErrDuplicatedConfig)
		}
	})
}

func TestReadSet(t *testing.T) {
	t.Run("Test a set can be read", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
}

func TestDeleteSet(t *testing.T) {
	t.Run("Test a set can be delete", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

		name := "mySet"
		service.CreateSet(name)

		_, err := service.DeleteSet("otherName")

		if err == nil {
			t.Errorf("Expected error: %v got nil", ports.ErrConfigNotExists)
		}
	})
}

func TestRenameSet(t *testing.T) {
	t.Run("Test a set can be renamed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
	t.Run("Test a new key can be added", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
	t.Run("Test a key can be updated", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
	t.Run("Test a key can be removed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
		service := NewConfigService(mockRepo, cacheRepo, &mockSecret)

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
