package main

import (
	"log"
	"net"
	"os"

	"flight-service/internal/cache"
	"flight-service/internal/db"
	"flight-service/internal/grpc"

	flightpb "flight-service/proto"

	grpcpkg "google.golang.org/grpc"
)

func main() {

	apiKey := os.Getenv("FLIGHT_API_KEY")
	if apiKey == "" {
		log.Fatal("FLIGHT_API_KEY environment variable is required")
	}

	database, err := db.Connect()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer database.Close()

	redisClient := cache.NewRedisClient()
	log.Println("Redis client initialized")

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpcpkg.NewServer(
		grpcpkg.UnaryInterceptor(grpc.AuthInterceptor()),
	)

	flightServer := grpc.NewFlightServer(database, redisClient)

	flightpb.RegisterFlightServiceServer(server, flightServer)

	log.Println("Flight Service gRPC server started on port 50051")

	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
