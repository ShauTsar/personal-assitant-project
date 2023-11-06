package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"log"
	"net"
	"personal-assitant-project/config"
	"personal-assitant-project/personal-assitant-server/grpc/handlers"
	pb "personal-assitant-project/personal-assitant-server/grpc/proto"
)

func main() {
	// Load database and Redis configurations from your configuration file
	dbConfig := config.LoadDatabaseConfig()
	//redisConfig := config.LoadRedisConfig()

	// Create a connection to PostgreSQL
	postgresDB, err := sql.Open("postgres", fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database, dbConfig.SSLMode,
	))
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	defer postgresDB.Close()
	// Create a Redis client
	//redisClient := redis.NewClient(&redis.Options{
	//	Addr: redisConfig.Addr,
	//	DB:   redisConfig.DB,
	//})

	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	server := grpc.NewServer()
	pb.RegisterUserServiceServer(server, &handlers.UserServiceHandler{})
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
