package main

import (
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"personal-assitant-project/personal-assitant-server/grpc/handlers"
	pb "personal-assitant-project/personal-assitant-server/grpc/proto/gen"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

func main() {

	flag.Parse()
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, &handlers.UserServiceServer{})
	reflection.Register(grpcServer)
	//go scanner.Scanner()
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
