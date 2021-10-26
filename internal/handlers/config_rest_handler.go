package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
	"golang.org/x/sync/singleflight"
)

// Feature flags used in ths handler
const (
	singleflightOn string = "single_flight_on"
)

type renameBody struct {
	Name string `json:"name"`
}

// ConfigRESTHandler provides a REST API handler for ports.ConfigService
type ConfigRESTHandler struct {
	config      *domain.Config
	service     ports.ConfigService
	toggleFlags ports.ToggleRepo
}

func NewConfigRESTHandler(
	config *domain.Config,
	toggleFlags ports.ToggleRepo,
	service ports.ConfigService) *ConfigRESTHandler {
	return &ConfigRESTHandler{
		config:      config,
		service:     service,
		toggleFlags: toggleFlags,
	}
}

// CreateRoutes adds the API routes to the gin router
func (handler *ConfigRESTHandler) CreateRoutes(router *gin.Engine) {

	// handler.config.APIPrefix
	group := router.Group("api")
	{
		group.GET("/config/:name", func(c *gin.Context) {
			var data []byte
			var err error
			if handler.toggleFlags.GetFlag(singleflightOn, context.Background()).Status {
				data, err = handler.getConfigJSONSingleFlight(c)
			} else {
				data, err = handler.GetConfigJSON(c)
			}

			if err != nil {
				handleError(err, c)
				return
			}

			c.Data(http.StatusOK, "application/json; charset=utf-8", data)
		})

		group.POST("/configset/:name/item", func(c *gin.Context) {
			data, err := handler.AddConfigItem(c)

			if err != nil {
				handleError(err, c)
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": data})
		})

		group.PATCH("/configset/:name/item", func(c *gin.Context) {
			data, err := handler.UpdateConfigItem(c)

			if err != nil {
				handleError(err, c)
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": data})
		})

		group.DELETE("/configset/:name/item/:key", func(c *gin.Context) {
			data, err := handler.DeleteConfigItem(c)

			if err != nil {
				handleError(err, c)
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": data})
		})

		group.GET("/configset/:name", func(c *gin.Context) {
			data, err := handler.GetConfigSet(c)
			if err != nil {
				handleError(err, c)
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": data})
		})

		group.POST("/configset/:name", func(c *gin.Context) {
			var data domain.ConfigSet
			var err error
			if c.Request.ContentLength > 0 {
				data, err = handler.RenameConfigSet(c)
			} else {
				data, err = handler.CreateConfigSet(c)
			}

			if err != nil {
				handleError(err, c)
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": data})
		})

		group.DELETE("/configset/:name", func(c *gin.Context) {
			data, err := handler.DeleteConfigSet(c)

			if err != nil {
				handleError(err, c)
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": data})
		})
	}
}

func (handler *ConfigRESTHandler) GetConfigJSON(c *gin.Context) ([]byte, error) {
	name, ok := c.Params.Get("name")

	if !ok {
		return nil, domain.ErrMissingParam("name")
	}

	ageQuery := c.Query("maxAge")
	age := domain.AnyAge
	if ageQuery != "" {
		var err error
		age, err = strconv.Atoi(ageQuery)

		if err != nil {
			return nil, domain.InvalidParam("maxAge")
		}
	}

	output, err := handler.service.GetSetJson(name, age)
	if err != nil {
		if err == ports.ErrConfigNotExists {
			return nil, domain.ErrNotFound(name)
		}

		log.Error().Stack().Err(err).Msg("GetConfigJSON error")
		return nil, &domain.ErrInternalError
	}

	output = append([]byte(`{"data":`), output...)
	output = append(output, []byte("}")...)
	return output, nil
}

func (handler *ConfigRESTHandler) CreateConfigSet(c *gin.Context) (domain.ConfigSet, error) {
	name, ok := c.Params.Get("name")

	if !ok {
		return domain.ConfigSet{}, domain.ErrMissingParam("name")
	}

	output, err := handler.service.CreateSet(name)
	if err != nil {
		if err == ports.ErrDuplicatedConfig {
			return domain.ConfigSet{}, domain.ErrBadRequest(err.Error())
		}

		log.Error().Stack().Err(err).Msg("GetConfigSet error")
		return domain.ConfigSet{}, &domain.ErrInternalError
	}

	return output, nil
}

func (handler *ConfigRESTHandler) RenameConfigSet(c *gin.Context) (domain.ConfigSet, error) {
	name, ok := c.Params.Get("name")

	if !ok {
		return domain.ConfigSet{}, domain.ErrMissingParam("name")
	}

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid body")
	}
	var body renameBody
	err = json.Unmarshal(jsonData, &body)
	if err != nil {
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid body")
	}

	if body.Name == "" || body.Name == name {
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid new name")
	}

	output, err := handler.service.RenameSet(name, body.Name)
	if err != nil {
		if err == ports.ErrConfigNotExists {
			return domain.ConfigSet{}, domain.ErrNotFound(name)
		}

		log.Error().Stack().Err(err).Msg("GetConfigSet error")
		return domain.ConfigSet{}, &domain.ErrInternalError
	}

	return output, nil
}

func (handler *ConfigRESTHandler) GetConfigSet(c *gin.Context) (domain.ConfigSet, error) {
	name, ok := c.Params.Get("name")

	if !ok {
		return domain.ConfigSet{}, domain.ErrMissingParam("name")
	}

	output, err := handler.service.GetSet(name)
	if err != nil {
		if err == ports.ErrConfigNotExists {
			return domain.ConfigSet{}, domain.ErrNotFound(name)
		}

		log.Error().Stack().Err(err).Msg("GetConfigSet error")
		return domain.ConfigSet{}, &domain.ErrInternalError
	}

	return output, nil
}

func (handler *ConfigRESTHandler) DeleteConfigSet(c *gin.Context) (domain.ConfigSet, error) {
	name, ok := c.Params.Get("name")

	if !ok {
		return domain.ConfigSet{}, domain.ErrMissingParam("name")
	}

	output, err := handler.service.DeleteSet(name)
	if err != nil {
		if err == ports.ErrConfigNotExists {
			return domain.ConfigSet{}, domain.ErrNotFound(name)
		}

		log.Error().Stack().Err(err).Msg("GetConfigSet error")
		return domain.ConfigSet{}, &domain.ErrInternalError
	}

	return output, nil
}

func (handler *ConfigRESTHandler) AddConfigItem(c *gin.Context) (domain.ConfigSet, error) {
	name, ok := c.Params.Get("name")

	if !ok {
		return domain.ConfigSet{}, domain.ErrMissingParam("name")
	}

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid body")
	}
	var body domain.ConfigItem
	err = json.Unmarshal(jsonData, &body)
	if err != nil {
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid body")
	}

	output, err := handler.service.AddItem(body, name)
	if err != nil {
		if err == domain.ErrDuplicatedKey {
			return domain.ConfigSet{}, domain.ErrBadRequest(err.Error())
		}

		log.Error().Stack().Err(err).Msg("GetConfigSet error")
		return domain.ConfigSet{}, &domain.ErrInternalError
	}

	return output, nil
}

func (handler *ConfigRESTHandler) UpdateConfigItem(c *gin.Context) (domain.ConfigSet, error) {
	name, ok := c.Params.Get("name")

	if !ok {
		return domain.ConfigSet{}, domain.ErrMissingParam("name")
	}

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid body")
	}
	var body domain.ConfigItem
	err = json.Unmarshal(jsonData, &body)
	if err != nil {
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid body")
	}

	output, err := handler.service.UpdateItem(body, name)
	if err != nil {
		if err == domain.ErrKeyNotExists {
			return domain.ConfigSet{}, domain.ErrNotFound(body.Key)
		}

		log.Error().Stack().Err(err).Msg("GetConfigSet error")
		return domain.ConfigSet{}, &domain.ErrInternalError
	}

	return output, nil
}

func (handler *ConfigRESTHandler) DeleteConfigItem(c *gin.Context) (domain.ConfigSet, error) {
	name, ok := c.Params.Get("name")

	if !ok {
		return domain.ConfigSet{}, domain.ErrMissingParam("name")
	}

	key, ok := c.Params.Get("key")

	if !ok {
		return domain.ConfigSet{}, domain.ErrMissingParam("key")
	}

	output, err := handler.service.RemoveItem(*domain.NewConfigItem(key, "", domain.Plain), name)
	if err != nil {
		if err == domain.ErrKeyNotExists {
			return domain.ConfigSet{}, domain.ErrNotFound(key)
		}

		log.Error().Stack().Err(err).Msg("GetConfigSet error")
		return domain.ConfigSet{}, &domain.ErrInternalError
	}

	return output, nil
}

// Single flight with channels and timeout
var getConfigJSONReqGroup singleflight.Group

func (handler *ConfigRESTHandler) getConfigJSONSingleFlight(c *gin.Context) ([]byte, error) {
	fp := fullPath(c)
	ch := getConfigJSONReqGroup.DoChan(fp, func() (interface{}, error) {
		return handler.GetConfigJSON(c)
	})

	// Create our timeout
	timeout := time.After(500 * time.Millisecond)

	var result singleflight.Result
	select {
	case <-timeout: // Timeout elapsed, send a timeout message (504)
		return nil, &domain.ErrTimeout
	case result = <-ch: // Received result from channel
	}

	// singleflight.Result is the same three values as returned from Do(), but wrapped
	// in a struct. Third return value tells if the output was shared to multiple callers
	if result.Err != nil {
		return nil, result.Err
	}

	return result.Val.([]byte), nil
}

// Utils

func handleError(err error, c *gin.Context) {
	log.Error().Stack().Err(err).Msg("Request error")

	if rest, ok := err.(*domain.RestError); ok {
		c.JSON(rest.HTTPStatus, gin.H{"error": rest})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func fullPath(c *gin.Context) string {
	fullPath := c.Request.URL.Path
	raw := c.Request.URL.RawQuery

	if raw != "" {
		fullPath = fullPath + "?" + raw
	}

	return fullPath
}
