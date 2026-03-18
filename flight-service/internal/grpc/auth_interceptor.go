package grpc

import (
	"context"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor проверяет API ключ для всех gRPC вызовов
func AuthInterceptor() grpc.UnaryServerInterceptor {

	apiKey := os.Getenv("FLIGHT_API_KEY")

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		// получаем metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// получаем ключ
		keys := md.Get("x-api-key")
		if len(keys) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing api key")
		}

		if keys[0] != apiKey {
			return nil, status.Error(codes.Unauthenticated, "invalid api key")
		}

		// если всё ок — выполняем метод
		return handler(ctx, req)
	}
}