package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	minervaLog "github.com/sy-software/minerva-go-utils/log"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/service"
	"github.com/sy-software/minerva-olive/internal/handlers"
	"github.com/sy-software/minerva-olive/internal/repositories/awssm"
	"github.com/sy-software/minerva-olive/internal/repositories/redis"
)

const defaultConfigFile = "./config.json"

func RequestId(c *gin.Context) {
	// Check for incoming header, use it if exists
	requestID := c.Request.Header.Get("X-Request-Id")

	// Create request id with UUID4
	if requestID == "" {
		uuid4, _ := uuid.NewV4()
		requestID = uuid4.String()
	}

	// Expose it for use in the application
	c.Set(domain.RequestIdKey, requestID)

	// Set X-Request-Id header
	c.Writer.Header().Set("X-REQUEST-ID", requestID)
	c.Next()
}

func main() {
	minervaLog.ConfigureLogger(minervaLog.LogLevel(os.Getenv("LOG_LEVEL")), os.Getenv("CONSOLE_OUTPUT") != "")
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.DurationFieldInteger = true
	log.Info().Msg("Starting server")

	config := domain.LoadConfig()
	db, err := redis.GetRedisDB(&config)
	if err != nil {
		log.Error().Stack().Err(err).Msg("Can't initialize Redis DB")
		os.Exit(1)
	}
	repo := redis.NewRedisRepo(&config, db)
	// TODO: Use a separated DB
	toggleRepo := redis.NewRedisToggleRepo(&config, db)
	if err != nil {
		log.Error().Stack().Err(err).Msg("Can't initialize Redis DB")
		os.Exit(1)
	}
	secretMngr := awssm.NewAWSSM()
	configService := service.NewConfigService(&config, repo, repo, secretMngr)

	handler := handlers.NewConfigRESTHandler(&config, toggleRepo, configService)

	router := gin.New()
	router.Use()
	router.Use(handlers.LogMiddleware("olive"))
	handler.CreateRoutes(router)

	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	srv := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		log.Info().Msgf("Server started at: %v", address)
		err := srv.ListenAndServe()

		if err != http.ErrServerClosed {
			log.Panic().Err(err).Msg("Server crashed")
		} else {
			log.Info().Msg("Server closed")
		}
	}()

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Info().Msg("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Panic().Err(err).Msg("Server forced to shutdown")
	}
}
