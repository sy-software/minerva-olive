package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/service"
	"github.com/sy-software/minerva-olive/mocks"
)

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestGetConfigJSON(t *testing.T) {
	t.Run("Test getting a config json", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		service.CreateSet(name)
		handler := NewConfigRESTHandler(&config, service)
		handler.CreateRoutes(router)

		got := performRequest(router, "GET", "/api/config/"+name)

		if got.Code != http.StatusOK {
			t.Errorf("Expected status code: %d, got: %d", http.StatusOK, got.Code)
		}

		expected := `{"data":{}}`
		if got.Body.String() != expected {
			t.Errorf("Expected config: %v got: %v", expected, got.Body.String())
		}
	})

	t.Run("Test getting a config json with max age", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		expectedMaxAge := 666
		called := false
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		cacheRepo.GetJSONInterceptor = func(key string, maxAge int) ([]byte, error) {
			called = true
			if maxAge != expectedMaxAge {
				t.Errorf("Expected max age: %d, got: %d", expectedMaxAge, maxAge)
			}
			return []byte("{}"), nil
		}
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		service.CreateSet(name)
		handler := NewConfigRESTHandler(&config, service)
		handler.CreateRoutes(router)

		got := performRequest(router, "GET", fmt.Sprintf("/api/config/%s?maxAge=%d", name, expectedMaxAge))

		if got.Code != http.StatusOK {
			t.Errorf("Expected status code: %d, got: %d", http.StatusOK, got.Code)
		}

		expected := `{"data":{}}`
		if got.Body.String() != expected {
			t.Errorf("Expected config: %v got: %v", expected, got.Body.String())
		}

		if !called {
			t.Errorf("Expected repo to be called")
		}
	})

	t.Run("Test getting a non existing config json", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		handler := NewConfigRESTHandler(&config, service)
		handler.CreateRoutes(router)

		got := performRequest(router, "GET", "/api/config/"+name)

		if got.Code != http.StatusNotFound {
			t.Errorf("Expected status code: %d, got: %d", http.StatusNotFound, got.Code)
		}

		expected := fmt.Sprint(domain.NotFound)
		if !strings.Contains(got.Body.String(), expected) {
			t.Errorf("Expected response to contain: %s got: %v", expected, got.Body.String())
		}
	})
}
