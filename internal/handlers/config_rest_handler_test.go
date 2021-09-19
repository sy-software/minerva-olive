package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
	"github.com/sy-software/minerva-olive/internal/core/service"
	"github.com/sy-software/minerva-olive/mocks"
)

func performRequest(r http.Handler, method, path string, body *string) *httptest.ResponseRecorder {
	var bodyReader io.Reader = nil
	if body != nil {
		bodyReader = strings.NewReader(*body)
	}

	req, _ := http.NewRequest(method, path, bodyReader)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

var toogleRepo = mocks.NewToggleFlagRepo(map[string]domain.ToggleFlag{
	singleflightOn: {
		Status: false,
	},
})

func TestMain(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
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
		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		got := performRequest(router, "GET", "/api/config/"+name, nil)

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
		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		got := performRequest(router, "GET", fmt.Sprintf("/api/config/%s?maxAge=%d", name, expectedMaxAge), nil)

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

		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		got := performRequest(router, "GET", "/api/config/"+name, nil)

		if got.Code != http.StatusNotFound {
			t.Errorf("Expected status code: %d, got: %d", http.StatusNotFound, got.Code)
		}

		expected := fmt.Sprint(domain.NotFound)
		if !strings.Contains(got.Body.String(), expected) {
			t.Errorf("Expected response to contain: %s got: %v", expected, got.Body.String())
		}
	})
}

func TestGetConfig(t *testing.T) {
	t.Run("Test getting a config for editing", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		service.CreateSet(name)
		set, _ := service.GetSet(name)
		setJson, _ := json.Marshal(gin.H{"data": set})
		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		got := performRequest(router, "GET", "/api/configset/"+name, nil)

		if got.Code != http.StatusOK {
			t.Errorf("Expected status code: %d, got: %d", http.StatusOK, got.Code)
		}

		expected := string(setJson)
		if got.Body.String() != expected {
			t.Errorf("Expected config: %v got: %v", expected, got.Body.String())
		}
	})

	t.Run("Test getting a non existing config", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		got := performRequest(router, "GET", "/api/configset/"+name, nil)

		if got.Code != http.StatusNotFound {
			t.Errorf("Expected status code: %d, got: %d", http.StatusNotFound, got.Code)
		}

		expected := fmt.Sprint(domain.NotFound)
		if !strings.Contains(got.Body.String(), expected) {
			t.Errorf("Expected response to contain: %s got: %v", expected, got.Body.String())
		}
	})
}

func TestCreateConfig(t *testing.T) {
	t.Run("Test create a config", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		got := performRequest(router, "POST", "/api/configset/"+name, nil)

		if got.Code != http.StatusOK {
			t.Errorf("Expected status code: %d, got: %d", http.StatusOK, got.Code)
		}

		var gotParsed struct {
			data domain.ConfigSet
		}
		err := json.Unmarshal(got.Body.Bytes(), &gotParsed)
		if err != nil {
			t.Errorf("Expected valid json got error: %v", err)
		}

		if gotParsed.data.Name == name {
			t.Errorf("Expected config with name: %v got: %v", name, gotParsed.data.Name)
		}
	})

	t.Run("Test create a duplicated config", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		service.CreateSet(name)
		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		got := performRequest(router, "POST", "/api/configset/"+name, nil)

		if got.Code != http.StatusBadRequest {
			t.Errorf("Expected status code: %d, got: %d", http.StatusBadRequest, got.Code)
		}

		expected := fmt.Sprint(ports.ErrDuplicatedConfig)
		if !strings.Contains(got.Body.String(), expected) {
			t.Errorf("Expected response to contain: %s got: %v", expected, got.Body.String())
		}
	})
}

func TestRenameConfig(t *testing.T) {
	t.Run("Test rename a config", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		service.CreateSet(name)
		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		newName := "newName"
		body := fmt.Sprintf(`{ "name": %q }`, newName)
		got := performRequest(router, "POST", "/api/configset/"+name, &body)

		if got.Code != http.StatusOK {
			t.Errorf("Expected status code: %d, got: %d", http.StatusOK, got.Code)
		}

		var gotParsed struct {
			data domain.ConfigSet
		}
		err := json.Unmarshal(got.Body.Bytes(), &gotParsed)
		if err != nil {
			t.Errorf("Expected valid json got error: %v", err)
		}

		if gotParsed.data.Name == newName {
			t.Errorf("Expected config with name: %v got: %v", name, gotParsed.data.Name)
		}
	})

	t.Run("Test rename a non existing config", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		newName := "newName"
		body := fmt.Sprintf(`{ "name": %q }`, newName)
		got := performRequest(router, "POST", "/api/configset/"+name, &body)

		if got.Code != http.StatusNotFound {
			t.Errorf("Expected status code: %d, got: %d", http.StatusNotFound, got.Code)
		}

		expected := fmt.Sprint(domain.NotFound)
		if !strings.Contains(got.Body.String(), expected) {
			t.Errorf("Expected response to contain: %s got: %v", expected, got.Body.String())
		}
	})
}

func TestDeleteConfig(t *testing.T) {
	t.Run("Test delete a config", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		service.CreateSet(name)
		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		got := performRequest(router, "DELETE", "/api/configset/"+name, nil)

		if got.Code != http.StatusOK {
			t.Errorf("Expected status code: %d, got: %d", http.StatusOK, got.Code)
		}

		got = performRequest(router, "GET", "/api/configset/"+name, nil)
		if got.Code != http.StatusNotFound {
			t.Errorf("Expected status code: %d, got: %d", http.StatusNotFound, got.Code)
		}
	})

	t.Run("Test rename a non existing config", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		got := performRequest(router, "DELETE", "/api/configset/"+name, nil)
		if got.Code != http.StatusNotFound {
			t.Errorf("Expected status code: %d, got: %d", http.StatusNotFound, got.Code)
		}
	})
}

func TestAddConfigItem(t *testing.T) {
	t.Run("Test add a config item", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		service.CreateSet(name)
		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		key := domain.NewConfigItem("myKey", float64(100), domain.Plain)
		bodyBytes, _ := json.Marshal(key)
		body := string(bodyBytes)
		got := performRequest(router, "POST", "/api/configset/"+name+"/item", &body)

		if got.Code != http.StatusOK {
			t.Errorf("Expected status code: %d, got: %d", http.StatusOK, got.Code)
		}

		var gotParsed struct {
			Data domain.ConfigSet
		}
		b := got.Body.Bytes()
		err := json.Unmarshal(b, &gotParsed)
		if err != nil {
			t.Errorf("Expected valid json got error: %v", err)
		}

		gotItem, err := gotParsed.Data.Get(key.Key)
		if err != nil {
			t.Errorf("Expected key item got error: %v", err)
		}
		if !cmp.Equal(*key, gotItem) {
			t.Errorf("Expected key item: %+v, got: %+v, diff: %v", key, gotItem, cmp.Diff(*key, gotItem))
		}
	})

	t.Run("Test add duplicated item", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		service.CreateSet(name)
		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		key := domain.NewConfigItem("myKey", 100, domain.Plain)
		bodyBytes, _ := json.Marshal(key)
		body := string(bodyBytes)
		performRequest(router, "POST", "/api/configset/"+name+"/item", &body)
		got := performRequest(router, "POST", "/api/configset/"+name+"/item", &body)

		if got.Code != http.StatusBadRequest {
			t.Errorf("Expected status code: %d, got: %d", http.StatusBadRequest, got.Code)
		}

		expected := fmt.Sprint(domain.ErrDuplicatedKey)
		if !strings.Contains(got.Body.String(), expected) {
			t.Errorf("Expected response to contain: %s got: %v", expected, got.Body.String())
		}
	})
}

func TestUpdateConfigItem(t *testing.T) {
	t.Run("Test update a config item", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		service.CreateSet(name)
		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		key := domain.NewConfigItem("myKey", float64(100), domain.Plain)
		bodyBytes, _ := json.Marshal(key)
		body := string(bodyBytes)
		performRequest(router, "POST", "/api/configset/"+name+"/item", &body)

		key.Value = float64(101)
		bodyBytes, _ = json.Marshal(key)
		body = string(bodyBytes)
		got := performRequest(router, "PATCH", "/api/configset/"+name+"/item", &body)

		if got.Code != http.StatusOK {
			t.Errorf("Expected status code: %d, got: %d", http.StatusOK, got.Code)
		}

		var gotParsed struct {
			Data domain.ConfigSet
		}
		b := got.Body.Bytes()
		err := json.Unmarshal(b, &gotParsed)
		if err != nil {
			t.Errorf("Expected valid json got error: %v", err)
		}

		gotItem, err := gotParsed.Data.Get(key.Key)
		if err != nil {
			t.Errorf("Expected key item got error: %v", err)
		}
		if !cmp.Equal(*key, gotItem) {
			t.Errorf("Expected key item: %+v, got: %+v, diff: %v", key, gotItem, cmp.Diff(*key, gotItem))
		}
	})

	t.Run("Test update a non existing config item", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		service.CreateSet(name)
		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		key := domain.NewConfigItem("myKey", float64(100), domain.Plain)
		bodyBytes, _ := json.Marshal(key)
		body := string(bodyBytes)
		performRequest(router, "POST", "/api/configset/"+name+"/item", &body)

		key.Key = "notexists"
		key.Value = float64(101)
		bodyBytes, _ = json.Marshal(key)
		body = string(bodyBytes)
		got := performRequest(router, "PATCH", "/api/configset/"+name+"/item", &body)

		if got.Code != http.StatusNotFound {
			t.Errorf("Expected status code: %d, got: %d", http.StatusNotFound, got.Code)
		}

		var gotParsed struct {
			Data domain.ConfigSet
		}
		b := got.Body.Bytes()
		err := json.Unmarshal(b, &gotParsed)
		if err != nil {
			t.Errorf("Expected valid json got error: %v", err)
		}

		_, err = gotParsed.Data.Get(key.Key)
		if err != domain.ErrKeyNotExists {
			t.Errorf("Expected error: %v got error: %v", domain.ErrKeyNotExists, err)
		}
	})
}

func TestDeleteConfigItem(t *testing.T) {
	t.Run("Test delete a config item", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		service.CreateSet(name)
		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		key := domain.NewConfigItem("myKey", float64(100), domain.Plain)
		bodyBytes, _ := json.Marshal(key)
		body := string(bodyBytes)
		performRequest(router, "POST", "/api/configset/"+name+"/item", &body)
		got := performRequest(router, "DELETE", "/api/configset/"+name+"/item/"+key.Key, nil)

		if got.Code != http.StatusOK {
			t.Errorf("Expected status code: %d, got: %d", http.StatusOK, got.Code)
		}

		var gotParsed struct {
			Data domain.ConfigSet
		}
		b := got.Body.Bytes()
		err := json.Unmarshal(b, &gotParsed)
		if err != nil {
			t.Errorf("Expected valid json got error: %v", err)
		}

		_, err = gotParsed.Data.Get(key.Key)
		if err != domain.ErrKeyNotExists {
			t.Errorf("Expected error: %v got error: %v", domain.ErrKeyNotExists, err)
		}
	})

	t.Run("Test delete a non existing config item", func(t *testing.T) {
		router := gin.New()
		name := "myConfig"
		config := domain.DefaultConfig()
		mockRepo := mocks.NewMockRepo()
		cacheRepo := mocks.NewMockRepo()
		mockSecret := mocks.MockSecrets{}
		service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

		service.CreateSet(name)
		handler := NewConfigRESTHandler(&config, toogleRepo, service)
		handler.CreateRoutes(router)

		key := domain.NewConfigItem("myKey", float64(100), domain.Plain)
		bodyBytes, _ := json.Marshal(key)
		body := string(bodyBytes)
		performRequest(router, "POST", "/api/configset/"+name+"/item", &body)
		got := performRequest(router, "DELETE", "/api/configset/"+name+"/item/notExists", nil)

		if got.Code != http.StatusNotFound {
			t.Errorf("Expected status code: %d, got: %d", http.StatusNotFound, got.Code)
		}
	})
}

// Benchmarks

func BenchmarkGetJSONWithoutSingleflight(b *testing.B) {
	router := gin.New()
	name := "myConfig"
	config := domain.DefaultConfig()
	mockRepo := mocks.NewMockRepo()
	cacheRepo := mocks.NewMockRepo()
	mockSecret := mocks.MockSecrets{}
	service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

	service.CreateSet(name)
	handler := NewConfigRESTHandler(&config, toogleRepo, service)
	handler.CreateRoutes(router)

	for i := 0; i < b.N; i++ {
		performRequest(router, "GET", "/api/config/"+name, nil)
	}
}

func BenchmarkGetJSONWithSingleflight(b *testing.B) {
	router := gin.New()
	name := "myConfig"
	config := domain.DefaultConfig()
	mockRepo := mocks.NewMockRepo()
	cacheRepo := mocks.NewMockRepo()
	mockSecret := mocks.MockSecrets{}
	service := service.NewConfigService(&config, mockRepo, cacheRepo, &mockSecret)

	service.CreateSet(name)
	toogleRepo := mocks.NewToggleFlagRepo(map[string]domain.ToggleFlag{
		singleflightOn: {
			Status: true,
		},
	})
	handler := NewConfigRESTHandler(&config, toogleRepo, service)
	handler.CreateRoutes(router)

	for i := 0; i < b.N; i++ {
		performRequest(router, "GET", "/api/config/"+name, nil)
	}
}
