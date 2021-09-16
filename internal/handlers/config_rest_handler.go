package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
)

type renameBody struct {
	Name string `json:"name"`
}

type ConfigRESTHandler struct {
	config  *domain.Config
	service ports.ConfigService
}

func NewConfigRESTHandler(config *domain.Config, service ports.ConfigService) *ConfigRESTHandler {
	return &ConfigRESTHandler{
		config:  config,
		service: service,
	}
}

func (handler *ConfigRESTHandler) CreateRoutes(router *gin.Engine) {

	// handler.config.APIPrefix
	group := router.Group("api")
	{
		group.GET("/config/:name", func(c *gin.Context) {
			data, err := handler.GetConfigJSON(c)

			if err != nil {
				handleError(err, c)
				return
			}

			// TODO: Remove this extra Unmarshal
			var out gin.H
			json.Unmarshal(data, &out)
			c.JSON(http.StatusOK, gin.H{"data": out})
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
			return domain.ConfigSet{}, domain.ErrBadRequest(err.Error(), domain.BadRequest)
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
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid body", domain.BadRequest)
	}
	var body renameBody
	err = json.Unmarshal(jsonData, &body)
	if err != nil {
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid body", domain.BadRequest)
	}

	if body.Name == "" || body.Name == name {
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid new name", domain.BadRequest)
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
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid body", domain.BadRequest)
	}
	var body domain.ConfigItem
	err = json.Unmarshal(jsonData, &body)
	if err != nil {
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid body", domain.BadRequest)
	}

	output, err := handler.service.AddItem(body, name)
	if err != nil {
		if err == domain.ErrDuplicatedKey {
			return domain.ConfigSet{}, domain.ErrBadRequest(err.Error(), domain.BadRequest)
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
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid body", domain.BadRequest)
	}
	var body domain.ConfigItem
	err = json.Unmarshal(jsonData, &body)
	if err != nil {
		return domain.ConfigSet{}, domain.ErrBadRequest("invalid body", domain.BadRequest)
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

// Utils

func handleError(err error, c *gin.Context) {
	log.Error().Stack().Err(err).Msg("Request error")

	if rest, ok := err.(*domain.RestError); ok {
		c.JSON(rest.HTTPStatus, gin.H{"error": rest})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
