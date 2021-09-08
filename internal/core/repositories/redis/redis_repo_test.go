package redis

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
)

// NOTE: redis-go/redis uses a mock redis server during tests

func TestSaveJSON(t *testing.T) {
	config := domain.DefaultConfig()
	repo, err := NewRedisRepo(&config)

	if err != nil {
		t.Errorf("Expected init without errors: %v", err)
	}

	expected := []byte("{}")
	err = repo.SaveJSON(expected, "mykey", domain.InfiniteTTL)

	if err != nil {
		t.Errorf("Expected save without errors: %v", err)
	}

	got, err := repo.GetJSON("mykey", domain.AnyAge)

	if err != nil {
		t.Errorf("Expected get without errors: %v", err)
	}

	if !cmp.Equal(got, expected) {
		t.Errorf("Expected: %q, got: %q", expected, got)
	}
}

func TestNotFoundError(t *testing.T) {
	config := domain.DefaultConfig()
	repo, err := NewRedisRepo(&config)

	if err != nil {
		t.Errorf("Expected init without errors: %v", err)
	}

	expected := []byte("{}")
	err = repo.SaveJSON(expected, "mykey", domain.InfiniteTTL)

	if err != nil {
		t.Errorf("Expected save without errors: %v", err)
	}

	_, err = repo.GetJSON("not_found", domain.AnyAge)

	if err != ports.ErrConfigNotExists {
		t.Errorf("Expected error: %v, got: %v", ports.ErrConfigNotExists, err)
	}
}

func TestCreateSet(t *testing.T) {
	config := domain.DefaultConfig()
	t.Run("Test a new config set can be created", func(t *testing.T) {
		repo, err := NewRedisRepo(&config)
		if err != nil {
			t.Errorf("Expected init without errors: %v", err)
		}

		name := "my_new_set"
		got, err := repo.CreateSet(*domain.NewConfigSet(name), domain.InfiniteTTL)

		if err != nil {
			t.Errorf("Expected set to be created without errors, got: %v", err)
		}

		if got.Name != name {
			t.Errorf("Expected name: %q, got %q", name, got.Name)
		}
	})

	t.Run("Test a duplicated config can't be created", func(t *testing.T) {
		repo, err := NewRedisRepo(&config)
		if err != nil {
			t.Errorf("Expected init without errors: %v", err)
		}

		name := "myset"
		repo.CreateSet(*domain.NewConfigSet(name), domain.InfiniteTTL)
		_, err = repo.CreateSet(*domain.NewConfigSet(name), domain.InfiniteTTL)

		if err != ports.ErrDuplicatedConfig {
			t.Errorf("Expected error: %v got nil", ports.ErrDuplicatedConfig)
		}
	})
}
