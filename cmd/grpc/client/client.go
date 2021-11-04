package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sy-software/minerva-olive/cmd/grpc/pb"
	"google.golang.org/grpc"
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	serverAddr := "localhost:8081"
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewConfigSetGRPCClient(conn)
	set, err := client.CreateConfigSet(context.Background(), &pb.NewConfigSet{Name: "helloworld2"})

	fmt.Printf("Set: %+v\n", set)
	fmt.Printf("Error: %q", err)
}
