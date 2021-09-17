package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sy-software/minerva-go-utils/datetime"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
)

const DB = 5

// Cache
func TestSaveGetJSON(t *testing.T) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, err := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())
	repo := NewRedisRepo(&config, db)

	if err != nil {
		t.Errorf("Expected init without errors: %v", err)
	}

	expected := []byte("{}")
	err = repo.SaveJSON(expected, "TestSaveJSON", domain.InfiniteTTL)

	if err != nil {
		t.Errorf("Expected save without errors: %v", err)
	}

	got, err := repo.GetJSON("TestSaveJSON", domain.AnyAge)

	if err != nil {
		t.Errorf("Expected get without errors: %v", err)
	}

	if !cmp.Equal(got, expected) {
		t.Errorf("Expected: %q, got: %q", expected, got)
	}
}

func TestRemoveJSON(t *testing.T) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, err := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())
	repo := NewRedisRepo(&config, db)

	if err != nil {
		t.Errorf("Expected init without errors: %v", err)
	}

	name := "TestSaveJSON"

	expected := []byte("{}")
	err = repo.SaveJSON(expected, name, domain.InfiniteTTL)

	if err != nil {
		t.Errorf("Expected save without errors: %v", err)
	}

	err = repo.RemoveJSON(name)
	if err != nil {
		t.Errorf("Expected remove without errors: %v", err)
	}

	_, err = repo.GetJSON(name, domain.AnyAge)
	if err != ports.ErrConfigNotExists {
		t.Errorf("Expected error: %v, got: %v", ports.ErrConfigNotExists, err)
	}
}

func TestNotFoundError(t *testing.T) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, err := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())
	repo := NewRedisRepo(&config, db)

	if err != nil {
		t.Errorf("Expected init without errors: %v", err)
	}

	expected := []byte("{}")
	err = repo.SaveJSON(expected, "TestNotFoundError", domain.InfiniteTTL)

	if err != nil {
		t.Errorf("Expected save without errors: %v", err)
	}

	_, err = repo.GetJSON("TestNotFoundErrorNotFound", domain.AnyAge)

	if err != ports.ErrConfigNotExists {
		t.Errorf("Expected error: %v, got: %v", ports.ErrConfigNotExists, err)
	}
}

// Persistent Config Sets

func TestCreateSet(t *testing.T) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, _ := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())
	t.Run("Test a new config set can be created", func(t *testing.T) {
		repo := NewRedisRepo(&config, db)

		name := "TestCreateSetOK"
		got, err := repo.CreateSet(*domain.NewConfigSet(name))

		if err != nil {
			t.Errorf("Expected set to be created without errors, got: %v", err)
		}

		if got.Name != name {
			t.Errorf("Expected name: %q, got %q", name, got.Name)
		}
	})

	t.Run("Test a duplicated config can't be created", func(t *testing.T) {
		repo := NewRedisRepo(&config, db)

		name := "TestCreateSetNotOK"
		repo.CreateSet(*domain.NewConfigSet(name))
		_, err := repo.CreateSet(*domain.NewConfigSet(name))

		if err != ports.ErrDuplicatedConfig {
			t.Errorf("Expected error: %v got nil", ports.ErrDuplicatedConfig)
		}
	})
}

func TestReadSet(t *testing.T) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, _ := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())
	t.Run("Test a set can be read", func(t *testing.T) {
		repo := NewRedisRepo(&config, db)

		name := "TestReadSetOK"
		repo.CreateSet(*domain.NewConfigSet(name))

		got, err := repo.GetSet(name)
		if err != nil {
			t.Errorf("Expected set to be read without errors, got: %v", err)
		}

		if got.Name != name {
			t.Errorf("Expected name: %q, got %q", name, got.Name)
		}
	})

	t.Run("Test a non existing set can't be read", func(t *testing.T) {
		repo := NewRedisRepo(&config, db)

		name := "TestReadSetNotFound"
		repo.CreateSet(*domain.NewConfigSet(name))

		_, err := repo.GetSet(name)
		if err != nil {
			t.Errorf("Expected error: %v got nil", ports.ErrConfigNotExists)
		}
	})

	t.Run("Test set names can be read and paginated", func(t *testing.T) {
		repo := NewRedisRepo(&config, db)
		db.Client.FlushDB(context.Background())

		for i := 0; i < 20; i++ {
			repo.CreateSet(*domain.NewConfigSet(fmt.Sprintf("TestReadSetPage%d", i)))
		}

		skip := 0
		count := 2
		names, err := repo.GetSetNames(count, skip)

		if err != nil {
			t.Errorf("Expected set names to be read without errors, got: %v", err)
		}

		if len(names) != count {
			t.Errorf("Expected names length of: %d, got: %d", count, len(names))
		}

		expected := []string{"set:TestReadSetPage0", "set:TestReadSetPage1"}
		if !cmp.Equal(names, expected) {
			t.Errorf("Expected names: %+v, got: %+v", expected, names)
		}

		skip = 20
		count = 2
		names, err = repo.GetSetNames(count, skip)

		if err != nil {
			t.Errorf("Expected set names to be read without errors, got: %v", err)
		}

		if len(names) != 0 {
			t.Errorf("Expected names length of: 0, got: %d", len(names))
		}
	})
}

func TestDeleteSet(t *testing.T) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, _ := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())

	t.Run("Test a set can be delete", func(t *testing.T) {
		repo := NewRedisRepo(&config, db)
		db.Client.FlushDB(context.Background())

		name := "TestDeleteSetOK"
		repo.CreateSet(*domain.NewConfigSet(name))

		_, err := repo.DeleteSet(name)
		if err != nil {
			t.Errorf("Expected set to be delete without errors, got: %v", err)
		}

		_, err = repo.GetSet(name)

		if err != ports.ErrConfigNotExists {
			t.Errorf("Expected error: %v got: %v", ports.ErrConfigNotExists, err)
		}
	})

	t.Run("Test a non existing set can't be deleted", func(t *testing.T) {
		repo := NewRedisRepo(&config, db)
		db.Client.FlushDB(context.Background())

		name := "TestDeleteSetOther"
		repo.CreateSet(*domain.NewConfigSet(name))

		_, err := repo.DeleteSet("TestDeleteSetNotFound")

		if err != ports.ErrConfigNotExists {
			t.Errorf("Expected error: %v got nil", ports.ErrConfigNotExists)
		}
	})
}

// Persistent ConfigItems

func TestAddItemToSet(t *testing.T) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, _ := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())

	t.Run("Test a new key can be added", func(t *testing.T) {
		repo := NewRedisRepo(&config, db)
		db.Client.FlushDB(context.Background())

		name := "TestAddItemToSetOK"
		repo.CreateSet(*domain.NewConfigSet(name))

		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)
		key := "TestAddItemToSetOKKey"
		newItem := domain.ConfigItem{
			Key:   key,
			Value: float64(100),
			Type:  domain.Plain,
		}
		_, err := repo.AddItem(newItem, name)
		if err != nil {
			t.Errorf("Expected item to be added without errors, got: %v", err)
		}

		got, err := repo.GetSet(name)
		if err != nil {
			t.Errorf("Expected item to be read without errors, got: %v", err)
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
			t.Errorf("Expected item: %+v, got: %+v, Diff: %+v", newItem, gotItem, cmp.Diff(gotItem, newItem))
		}
	})

	t.Run("Test a duplicated key can't be added", func(t *testing.T) {
		repo := NewRedisRepo(&config, db)
		db.Client.FlushDB(context.Background())

		name := "TestAddItemToSetDuplicated"
		repo.CreateSet(*domain.NewConfigSet(name))

		key := "TestAddItemToSetDuplicatedKey"
		newItem := domain.ConfigItem{
			Key:   key,
			Value: "100",
			Type:  domain.Plain,
		}
		repo.AddItem(newItem, name)
		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)
		now := datetime.UnixUTCNow()
		newItem.Value = "101"
		_, err := repo.AddItem(newItem, name)
		if err == nil {
			t.Errorf("Expected error: %v, got nil", domain.ErrDuplicatedKey)
		}

		got, err := repo.GetSet(name)
		if err != nil {
			t.Errorf("Expected item to be read without errors, got: %v", err)
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

		if gotItem.Value != "100" {
			t.Errorf("Expected item value: %+v, got: %+v", "100", gotItem.Value)
		}
	})
}

func TestUpdateItemFromSet(t *testing.T) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, _ := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())

	t.Run("Test a key can be updated", func(t *testing.T) {
		repo := NewRedisRepo(&config, db)
		db.Client.FlushDB(context.Background())

		name := "TestUpdateItemFromSetOK"
		repo.CreateSet(*domain.NewConfigSet(name))

		key := "TestUpdateItemFromSetOKKey"
		newItem := domain.ConfigItem{
			Key:   key,
			Value: "100",
			Type:  domain.Plain,
		}
		repo.AddItem(newItem, name)

		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)

		newItem.Value = "1000"
		_, err := repo.UpdateItem(newItem, name)
		if err != nil {
			t.Errorf("Expected item to be updated without errors, got: %v", err)
		}

		got, err := repo.GetSet(name)
		if err != nil {
			t.Errorf("Expected item to be read without errors, got: %v", err)
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
		repo := NewRedisRepo(&config, db)
		db.Client.FlushDB(context.Background())

		name := "TestUpdateItemFromSetNotFound"
		repo.CreateSet(*domain.NewConfigSet(name))

		key := "TestUpdateItemFromSetNotFoundKey"
		newItem := domain.ConfigItem{
			Key:   key,
			Value: "100",
			Type:  domain.Plain,
		}
		repo.AddItem(newItem, name)
		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)
		now := datetime.UnixUTCNow()
		newItem.Key = "TestUpdateItemFromSetNotFoundKeyOther"
		newItem.Value = "101"
		_, err := repo.UpdateItem(newItem, name)
		if err == nil {
			t.Errorf("Expected error: %v, got nil", domain.ErrKeyNotExists)
		}

		got, err := repo.GetSet(name)
		if err != nil {
			t.Errorf("Expected item to be read without errors, got: %v", err)
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

		if gotItem.Value != "100" {
			t.Errorf("Expected item value: %+v, got: %+v", "100", gotItem.Value)
		}
	})
}

func TestRemoveItemFromSet(t *testing.T) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, _ := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())
	t.Run("Test a key can be removed", func(t *testing.T) {
		repo := NewRedisRepo(&config, db)
		db.Client.FlushDB(context.Background())

		name := "TestRemoveItemFromSetOK"
		repo.CreateSet(*domain.NewConfigSet(name))

		key := "TestRemoveItemFromSetOKKey"
		newItem := domain.ConfigItem{
			Key:   key,
			Value: "100",
			Type:  domain.Plain,
		}
		repo.AddItem(newItem, name)

		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)

		newItem.Value = "1000"
		_, err := repo.RemoveItem(newItem, name)
		if err != nil {
			t.Errorf("Expected item to be removed without errors, got: %v", err)
		}

		got, err := repo.GetSet(name)
		if err != nil {
			t.Errorf("Expected item to be read without errors, got: %v", err)
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
		repo := NewRedisRepo(&config, db)
		db.Client.FlushDB(context.Background())

		name := "TestRemoveItemFromSetNotFound"
		repo.CreateSet(*domain.NewConfigSet(name))

		key := "TestRemoveItemFromSetNotFoundKey"
		newItem := domain.ConfigItem{
			Key:   key,
			Value: "100",
			Type:  domain.Plain,
		}
		repo.AddItem(newItem, name)
		// Wait a few seconds to have a token with different expiration
		time.Sleep(1 * time.Second)
		now := datetime.UnixUTCNow()
		newItem.Key = "TestRemoveItemFromSet_NotFound_Key"
		newItem.Value = "101"
		_, err := repo.RemoveItem(newItem, name)
		if err != domain.ErrKeyNotExists {
			t.Errorf("Expected error: %v, got nil", domain.ErrKeyNotExists)
		}

		got, err := repo.GetSet(name)
		if err != nil {
			t.Errorf("Expected item to be read without errors, got: %v", err)
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

		if gotItem.Value != "100" {
			t.Errorf("Expected item value: %+v, got: %+v", "100", gotItem.Value)
		}
	})
}

// Benchmarks

func BenchmarkGetJSON(b *testing.B) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, err := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())
	repo := NewRedisRepo(&config, db)

	if err != nil {
		b.Errorf("Expected init without errors: %v", err)
	}

	expected := []byte("{}")
	err = repo.SaveJSON(expected, "TestSaveJSON", domain.InfiniteTTL)

	if err != nil {
		b.Errorf("Expected save without errors: %v", err)
	}

	for i := 0; i < b.N; i++ {
		repo.GetJSON("TestSaveJSON", domain.AnyAge)
	}
}
