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

    client := grpcclient.NewFlightClient()

    grpcCtx := grpcclient.WithAPIKey(ctx)

    _, err := client.GetFlight(grpcCtx, &flightpb.GetFlightRequest{
        Id: int32(flightID),
    })

    if err != nil {
        return fmt.Errorf("flight service unavailable: %w", err)
    }

    return s.repo.CreateBooking(ctx, flightID, seats, name, email)
}
