package repository

import (
    "context"
    "database/sql"
    "fmt"

    _ "github.com/lib/pq"
)

type BookingRepository struct {
    db *sql.DB
}

func NewPostgresDB() (*sql.DB, error) {
    connStr := "postgres://postgres:postgres@booking-db:5432/bookingdb?sslmode=disable"
    return sql.Open("postgres", connStr)
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
    return &BookingRepository{db: db}
}

func (r *BookingRepository) CreateBooking(
    ctx context.Context,
    flightID int64,
    seats int32,
    name string,
    email string,
) error {

    query := `
        INSERT INTO bookings (flight_id, seats, passenger_name, passenger_email)
        VALUES ($1, $2, $3, $4)
    `

    _, err := r.db.ExecContext(ctx, query, flightID, seats, name, email)
    if err != nil {
        return fmt.Errorf("failed to insert booking: %w", err)
    }

    return nil
}