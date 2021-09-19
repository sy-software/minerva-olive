package redis

import (
	"context"
	"testing"

	"github.com/sy-software/minerva-olive/internal/core/domain"
)

const ToggleDB = 5

func TestGetSetToggleFlag(t *testing.T) {
	config := domain.DefaultConfig()
	config.Redis.DB = DB
	db, err := GetRedisDB(&config)
	defer db.Client.FlushDB(context.Background())
	repo := NewRedisToggleRepo(&config, db)

	if err != nil {
		t.Errorf("Expected init without errors: %v", err)
	}

	t.Run("Test setting and getting a flag value", func(t *testing.T) {
		flagName := "basic_flag"
		status := true
		err := repo.SetFlag(flagName, status, nil)
		if err != nil {
			t.Errorf("Expected setting a flag without errors, got: %v", err)
		}

		flag := repo.GetFlag(flagName, context.TODO())

		if flag.Status != status {
			t.Errorf("Expected flag status: %v got: %v", status, flag.Status)
		}
	})

	t.Run("Test getting non existing flag returns default", func(t *testing.T) {
		flagName := "no_existing_flag"
		status := false
		flag := repo.GetFlag(flagName, context.TODO())

		if flag.Status != status {
			t.Errorf("Expected flag status: %v got: %v", status, flag.Status)
		}
	})

	t.Run("Test overriding non existing flag default", func(t *testing.T) {
		flagName := "no_existing_flag"
		status := true
		flag := repo.GetFlagWithDefaults(flagName, status, nil, context.TODO())

		if flag.Status != status {
			t.Errorf("Expected flag status: %v got: %v", status, flag.Status)
		}
	})

	t.Run("Test updating a flag value", func(t *testing.T) {
		flagName := "basic_flag"
		status := true
		err := repo.SetFlag(flagName, status, nil)
		if err != nil {
			t.Errorf("Expected setting a flag without errors, got: %v", err)
		}

		flag := repo.GetFlag(flagName, context.TODO())

		if flag.Status != status {
			t.Errorf("Expected flag status: %v got: %v", status, flag.Status)
		}

		err = repo.SetFlag(flagName, !status, nil)
		if err != nil {
			t.Errorf("Expected updating a flag without errors, got: %v", err)
		}

		flag = repo.GetFlag(flagName, context.TODO())
		if flag.Status != !status {
			t.Errorf("Expected flag status: %v got: %v", !status, flag.Status)
		}
	})
}
