package redis

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
)

const DB = 5

func TestSaveJSON(t *testing.T) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, _ := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())
	repo, err := NewRedisRepo(&config, db)

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

func TestNotFoundError(t *testing.T) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, _ := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())
	repo, err := NewRedisRepo(&config, db)

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

func TestCreateSet(t *testing.T) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, _ := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())
	t.Run("Test a new config set can be created", func(t *testing.T) {
		repo, err := NewRedisRepo(&config, db)
		if err != nil {
			t.Errorf("Expected init without errors: %v", err)
		}

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
		repo, err := NewRedisRepo(&config, db)
		if err != nil {
			t.Errorf("Expected init without errors: %v", err)
		}

		name := "TestCreateSetNotOK"
		repo.CreateSet(*domain.NewConfigSet(name))
		_, err = repo.CreateSet(*domain.NewConfigSet(name))

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
		repo, err := NewRedisRepo(&config, db)
		if err != nil {
			t.Errorf("Expected init without errors: %v", err)
		}

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
		repo, err := NewRedisRepo(&config, db)
		if err != nil {
			t.Errorf("Expected init without errors: %v", err)
		}

		name := "TestReadSetNotFound"
		repo.CreateSet(*domain.NewConfigSet(name))

		_, err = repo.GetSet(name)
		if err != nil {
			t.Errorf("Expected error: %v got nil", ports.ErrConfigNotExists)
		}
	})

	t.Run("Test set names can be read and paginated", func(t *testing.T) {
		repo, err := NewRedisRepo(&config, db)
		db.Client.FlushDB(context.Background())
		if err != nil {
			t.Errorf("Expected init without errors: %v", err)
		}

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
