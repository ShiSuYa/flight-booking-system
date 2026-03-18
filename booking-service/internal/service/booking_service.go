package service

import (
    "context"
    "fmt"

    "booking-service/internal/grpcclient"
    "booking-service/internal/repository"
    flightpb "booking-service/proto"
)

type BookingService struct {
    repo *repository.BookingRepository
}

func NewBookingService(r *repository.BookingRepository) *BookingService {
    return &BookingService{repo: r}
}

func (s *BookingService) CreateBooking(
    ctx context.Context,
    flightID int64,
    seats int32,
    name string,
    email string,
) error {

    // 🔥 1. создаём gRPC клиент
    client := grpcclient.NewFlightClient()

    // 🔥 2. добавляем API key
    grpcCtx := grpcclient.WithAPIKey(ctx)

    // 🔥 3. ВАЖНО — вызываем Flight Service
    _, err := client.GetFlight(grpcCtx, &flightpb.GetFlightRequest{
        Id: int32(flightID),
    })

    if err != nil {
        return fmt.Errorf("flight service unavailable: %w", err)
    }

    // 🔥 4. если всё ок — создаём бронь
    return s.repo.CreateBooking(ctx, flightID, seats, name, email)
}