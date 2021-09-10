package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
)

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
		group.POST("/config", func(c *gin.Context) {
			data, err := handler.GetConfigJSON(c)

			if err != nil {
				handleError(err, c)
				return
			}

			c.JSON(http.StatusOK, gin.H{"data": data})
		})

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

// Utils

func handleError(err error, c *gin.Context) {
	log.Error().Stack().Err(err).Msg("Request error")

	if rest, ok := err.(*domain.RestError); ok {
		c.JSON(rest.HTTPStatus, gin.H{"error": rest})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
