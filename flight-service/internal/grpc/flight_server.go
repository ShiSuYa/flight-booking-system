package grpc

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	flightpb "flight-service/proto"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FlightServer struct {
	flightpb.UnimplementedFlightServiceServer
	db  *sql.DB
	rdb *redis.Client
}

func NewFlightServer(db *sql.DB, rdb *redis.Client) *FlightServer {
	return &FlightServer{
		db:  db,
		rdb: rdb,
	}
}

func authenticate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	keys := md.Get("x-api-key")
	if len(keys) == 0 {
		return status.Error(codes.Unauthenticated, "missing API key")
	}

	expected := os.Getenv("FLIGHT_API_KEY")
	if expected == "" {
		return status.Error(codes.Internal, "server API key not configured")
	}

	if keys[0] != expected {
		return status.Error(codes.Unauthenticated, "invalid API key")
	}

	return nil
}

func (s *FlightServer) cacheGet(ctx context.Context, key string) ([]byte, bool) {
	val, err := s.rdb.Get(ctx, key).Bytes()

	if err == redis.Nil {
		fmt.Println("Cache miss:", key)
		return nil, false
	}

	if err != nil {
		fmt.Println("Redis error:", err)
		return nil, false
	}

	fmt.Println("Cache hit:", key)
	return val, true
}

func (s *FlightServer) cacheSet(ctx context.Context, key string, value []byte, ttl time.Duration) {
	err := s.rdb.Set(ctx, key, value, ttl).Err()
	if err != nil {
		fmt.Println("Redis set error:", err)
	}
}

func (s *FlightServer) invalidateSearchCache(ctx context.Context) {
	iter := s.rdb.Scan(ctx, 0, "search:*", 0).Iterator()

	for iter.Next(ctx) {
		s.rdb.Del(ctx, iter.Val())
	}

	if err := iter.Err(); err != nil {
		fmt.Println("Redis scan error:", err)
	}
}

func (s *FlightServer) GetFlight(ctx context.Context, req *flightpb.GetFlightRequest) (*flightpb.GetFlightResponse, error) {

	if err := authenticate(ctx); err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("flight:%d", req.FlightId)

	if data, ok := s.cacheGet(ctx, cacheKey); ok {

		var resp flightpb.GetFlightResponse

		if err := proto.Unmarshal(data, &resp); err == nil {
			return &resp, nil
		}
	}

	query := `
	SELECT id, flight_number, origin, destination,
	       departure_time, arrival_time,
	       total_seats, available_seats, price, status
	FROM flights
	WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, req.FlightId)

	var (
		id, statusInt                   int64
		flightNumber, origin, destination string
		departureTime, arrivalTime       time.Time
		totalSeats, availableSeats       int32
		price                            float64
	)

	err := row.Scan(
		&id,
		&flightNumber,
		&origin,
		&destination,
		&departureTime,
		&arrivalTime,
		&totalSeats,
		&availableSeats,
		&price,
		&statusInt,
	)

	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "flight not found")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	flight := &flightpb.Flight{
		Id:             id,
		FlightNumber:   flightNumber,
		Origin:         origin,
		Destination:    destination,
		DepartureTime:  timestamppb.New(departureTime),
		ArrivalTime:    timestamppb.New(arrivalTime),
		TotalSeats:     totalSeats,
		AvailableSeats: availableSeats,
		Price:          price,
		Status:         flightpb.FlightStatus(statusInt),
	}

	resp := &flightpb.GetFlightResponse{
		Flight: flight,
	}

	data, _ := proto.Marshal(resp)

	s.cacheSet(ctx, cacheKey, data, 5*time.Minute)

	return resp, nil
}

func (s *FlightServer) SearchFlights(ctx context.Context, req *flightpb.SearchFlightsRequest) (*flightpb.SearchFlightsResponse, error) {

	if err := authenticate(ctx); err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf(
		"search:%s:%s:%s",
		req.Origin,
		req.Destination,
		req.DepartureDate.AsTime().Format("2006-01-02"),
	)

	if data, ok := s.cacheGet(ctx, cacheKey); ok {

		var resp flightpb.SearchFlightsResponse

		if err := proto.Unmarshal(data, &resp); err == nil {
			return &resp, nil
		}
	}

	query := `
	SELECT id, flight_number, origin, destination,
	       departure_time, arrival_time,
	       total_seats, available_seats, price, status
	FROM flights
	WHERE origin = $1
	AND destination = $2
	AND DATE(departure_time) = DATE($3)
	AND status = 1
	`

	rows, err := s.db.QueryContext(
		ctx,
		query,
		req.Origin,
		req.Destination,
		req.DepartureDate.AsTime(),
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	defer rows.Close()

	flights := []*flightpb.Flight{}

	for rows.Next() {

		var (
			id, statusInt                   int64
			flightNumber, origin, destination string
			departureTime, arrivalTime       time.Time
			totalSeats, availableSeats       int32
			price                            float64
		)

		err := rows.Scan(
			&id,
			&flightNumber,
			&origin,
			&destination,
			&departureTime,
			&arrivalTime,
			&totalSeats,
			&availableSeats,
			&price,
			&statusInt,
		)

		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		flights = append(flights, &flightpb.Flight{
			Id:             id,
			FlightNumber:   flightNumber,
			Origin:         origin,
			Destination:    destination,
			DepartureTime:  timestamppb.New(departureTime),
			ArrivalTime:    timestamppb.New(arrivalTime),
			TotalSeats:     totalSeats,
			AvailableSeats: availableSeats,
			Price:          price,
			Status:         flightpb.FlightStatus(statusInt),
		})
	}

	resp := &flightpb.SearchFlightsResponse{
		Flights: flights,
	}

	data, _ := proto.Marshal(resp)

	s.cacheSet(ctx, cacheKey, data, 5*time.Minute)

	return resp, nil
}

func (s *FlightServer) ReserveSeats(ctx context.Context, req *flightpb.ReserveSeatsRequest) (*flightpb.ReserveSeatsResponse, error) {

	if err := authenticate(ctx); err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	defer tx.Rollback()

	var available int32

	err = tx.QueryRowContext(
		ctx,
		"SELECT available_seats FROM flights WHERE id=$1 FOR UPDATE",
		req.FlightId,
	).Scan(&available)

	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "flight not found")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if available < req.Seats {
		return nil, status.Error(codes.ResourceExhausted, "not enough available seats")
	}

	_, err = tx.ExecContext(
		ctx,
		"UPDATE flights SET available_seats = available_seats - $1 WHERE id=$2",
		req.Seats,
		req.FlightId,
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO seat_reservations (flight_id, booking_id, seats_reserved, status, created_at) VALUES ($1,$2,$3,'ACTIVE',NOW())",
		req.FlightId,
		req.ReservationId,
		req.Seats,
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.rdb.Del(ctx, fmt.Sprintf("flight:%d", req.FlightId))
	s.invalidateSearchCache(ctx)

	return &flightpb.ReserveSeatsResponse{
		Success:        true,
		RemainingSeats: available - req.Seats,
	}, nil
}

func (s *FlightServer) ReleaseReservation(ctx context.Context, req *flightpb.ReleaseReservationRequest) (*flightpb.ReleaseReservationResponse, error) {

	if err := authenticate(ctx); err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	defer tx.Rollback()

	var seats int32

	err = tx.QueryRowContext(
		ctx,
		"SELECT seats_reserved FROM seat_reservations WHERE booking_id=$1 AND flight_id=$2 AND status='ACTIVE' FOR UPDATE",
		req.ReservationId,
		req.FlightId,
	).Scan(&seats)

	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "reservation not found")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = tx.ExecContext(
		ctx,
		"UPDATE seat_reservations SET status='RELEASED' WHERE booking_id=$1 AND flight_id=$2",
		req.ReservationId,
		req.FlightId,
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = tx.ExecContext(
		ctx,
		"UPDATE flights SET available_seats = available_seats + $1 WHERE id=$2",
		seats,
		req.FlightId,
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.rdb.Del(ctx, fmt.Sprintf("flight:%d", req.FlightId))
	s.invalidateSearchCache(ctx)

	return &flightpb.ReleaseReservationResponse{
		Success:        true,
		AvailableSeats: seats,
	}, nil
}