package service

import (
	"testing"

	"github.com/sy-software/minerva-go-utils/datetime"
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
		mockRepo.CreateSet(name, ports.InfiniteTTL)
		_, err := service.CreateSet(name)

		if err == nil {
			t.Errorf("Expected error: %v got nil", ports.ErrDuplicatedConfig)
		}
	})
}
