package handlers

import (
	"context"

	"github.com/sy-software/minerva-olive/cmd/grpc/pb"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/ports"
)

type ConfigGRPCHandler struct {
	pb.UnimplementedConfigSetGRPCServer

	config      *domain.Config
	service     ports.ConfigService
	toggleFlags ports.ToggleRepo
}

func NewConfigGRPCHandler(
	config *domain.Config,
	toggleFlags ports.ToggleRepo,
	service ports.ConfigService) *ConfigGRPCHandler {
	return &ConfigGRPCHandler{
		config:      config,
		service:     service,
		toggleFlags: toggleFlags,
	}
}

func (handler ConfigGRPCHandler) CreateConfigSet(ctx context.Context, newSet *pb.NewConfigSet) (*pb.ConfigSet, error) {
	set, err := handler.service.CreateSet(newSet.Name)

	if err != nil {
		return nil, err
	}

	return &pb.ConfigSet{
		Name: set.Name,
	}, nil
}
