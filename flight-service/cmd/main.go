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

	// читаем API ключ из переменной окружения
	apiKey := os.Getenv("FLIGHT_API_KEY")
	if apiKey == "" {
		log.Fatal("FLIGHT_API_KEY environment variable is required")
	}

	// подключение к БД (PostgreSQL)
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer database.Close()

	// подключение к Redis
	redisClient := cache.NewRedisClient()
	log.Println("Redis client initialized")

	// запуск gRPC listener
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// создаём gRPC сервер с interceptor для аутентификации
	server := grpcpkg.NewServer(
		grpcpkg.UnaryInterceptor(grpc.AuthInterceptor()),
	)

	// создаём FlightServer (передаём БД и Redis)
	flightServer := grpc.NewFlightServer(database, redisClient)

	// регистрируем gRPC сервис
	flightpb.RegisterFlightServiceServer(server, flightServer)

	log.Println("Flight Service gRPC server started on port 50051")

	// запускаем сервер
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}