package grpcclient

import (
	"context"
	"log"
	"os"

	flightpb "booking-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func NewFlightClient() flightpb.FlightServiceClient {

	cb := NewCircuitBreaker()

	conn, err := grpc.Dial(
		"flight-service:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(cb.UnaryClientInterceptor()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to Flight Service: %v", err)
	}

	return flightpb.NewFlightServiceClient(conn)
}

func WithAPIKey(ctx context.Context) context.Context {
	apiKey := os.Getenv("FLIGHT_SERVICE_API_KEY")

	md := metadata.Pairs("x-api-key", apiKey)
	return metadata.NewOutgoingContext(ctx, md)
}
