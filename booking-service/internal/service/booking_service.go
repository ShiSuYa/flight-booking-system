package service

import (
 "context"
 "fmt"
 "time"

 "booking-service/internal/grpcclient"
 "booking-service/internal/repository"
 flightpb "booking-service/proto"
)

type BookingService struct {
 repo         *repository.BookingRepository
 flightClient flightpb.FlightServiceClient
}

type CircuitBreaker struct {
 state        string
 failures     int
 lastAttempt  time.Time
 openTimeout  time.Duration
 maxFailures  int
}

func NewCircuitBreaker(maxFailures int, openTimeout time.Duration) *CircuitBreaker {
 return &CircuitBreaker{
  state:       "closed",
  failures:    0,
  openTimeout: openTimeout,
  maxFailures: maxFailures,
 }
}

func (cb *CircuitBreaker) AllowRequest() bool {
 if cb.state == "open" {
  if time.Since(cb.lastAttempt) > cb.openTimeout {
   cb.state = "half"
   return true
  }
  return false
 }
 return true
}

func (cb *CircuitBreaker) Success() {
 cb.failures = 0
 if cb.state == "half" {
  cb.state = "closed"
 }
}

func (cb *CircuitBreaker) Failure() {
 cb.failures++
 cb.lastAttempt = time.Now()
 if cb.failures >= cb.maxFailures {
  cb.state = "open"
 }
}

func NewBookingService(r *repository.BookingRepository, fc flightpb.FlightServiceClient) *BookingService {
 return &BookingService{repo: r, flightClient: fc}
}

func (s *BookingService) CreateBooking(ctx context.Context, flightID int64, seats int32, name, email string, totalPrice float64, status string, cb *CircuitBreaker) error {

 if !cb.AllowRequest() {
  return fmt.Errorf("flight service unavailable (circuit breaker open)")
 }

 var lastErr error
 for i := 0; i < 3; i++ {
  _, err := s.flightClient.GetFlight(ctx, &flightpb.GetFlightRequest{
   FlightId: flightID,
  })
  if err != nil {
   lastErr = err
   cb.Failure()
   time.Sleep(100 * time.Millisecond)
   continue
  }
  cb.Success()
  lastErr = nil
  break
 }

 if lastErr != nil {
  return fmt.Errorf("flight service unavailable: %w", lastErr)
 }

 return s.repo.CreateBooking(ctx, flightID, seats, name, email, totalPrice, status)
}

func (s *BookingService) CancelBooking(ctx context.Context, bookingID int64) error {
 return s.repo.CancelBooking(ctx, bookingID)
}

func (s *BookingService) GetAllBookings(ctx context.Context) ([]repository.Booking, error) {
 return s.repo.GetAllBookings(ctx)
}

func (s *BookingService) GetBookingByID(ctx context.Context, id int64) (*repository.Booking, error) {
 return s.repo.GetBookingByID(ctx, id)
}