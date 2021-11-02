package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/sy-software/minerva-olive/cmd/grpc/pb"
	"github.com/sy-software/minerva-olive/internal/core/domain"
	"github.com/sy-software/minerva-olive/internal/core/service"
	"github.com/sy-software/minerva-olive/internal/repositories/awssm"
	"github.com/sy-software/minerva-olive/internal/repositories/redis"
	grpc "google.golang.org/grpc"
)

type gRPCServer struct {
	pb.UnimplementedConfigSetGRPCServer
}

func (gRPCServer) CreateConfigSet(ctx context.Context, newSet *pb.NewConfigSet) (*pb.ConfigSet, error) {
	config := domain.LoadConfig()
	db, err := redis.GetRedisDB(&config)
	if err != nil {
		log.Fatalf("Can't initialize Redis DB")
		os.Exit(1)
	}
	repo := redis.NewRedisRepo(&config, db)
	// TODO: Use a separated DB
	if err != nil {
		log.Fatalf("Can't initialize Redis DB")
		os.Exit(1)
	}
	secretMngr := awssm.NewAWSSM()
	configService := service.NewConfigService(&config, repo, repo, secretMngr)

	set, err := configService.CreateSet(newSet.Name)

	if err != nil {
		return nil, err
	}

	return &pb.ConfigSet{
		Name: set.Name,
	}, nil
}

func main() {
	flag.Parse()
	port := flag.Int("port", 8080, "The server port")
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterConfigSetGRPCServer(grpcServer, &gRPCServer{})
	grpcServer.Serve(lis)
}
