package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	minervaLog "github.com/sy-software/minerva-go-utils/log"
	"github.com/sy-software/minerva-olive/cmd/grpc/pb"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/service"
	"github.com/sy-software/minerva-olive/internal/handlers"
	"github.com/sy-software/minerva-olive/internal/repositories/awssm"
	"github.com/sy-software/minerva-olive/internal/repositories/redis"
	grpc "google.golang.org/grpc"
)

func main() {
	minervaLog.ConfigureLogger(minervaLog.LogLevel(os.Getenv("LOG_LEVEL")), os.Getenv("CONSOLE_OUTPUT") != "")
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.DurationFieldInteger = true
	log.Info().Msg("Starting gRPC server")

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

	handler := handlers.NewConfigGRPCHandler(&config, toggleRepo, configService)

	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	log.Info().Msgf("gRPC Server Listen at: %v", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start gRPC")
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterConfigSetGRPCServer(grpcServer, handler)
	grpcServer.Serve(lis)
}
