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

type Booking struct {
 ID            int64
 FlightID      int64
 Seats         int32
 PassengerName string
 PassengerEmail string
 TotalPrice    float64
 Status        string
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
 name, email string,
 totalPrice float64,
 status string,
) error {

 tx, err := r.db.BeginTx(ctx, nil)
 if err != nil {
  return err
 }

 query := `
        INSERT INTO bookings (
            flight_id, seats, passenger_name, passenger_email, total_price, status
        ) VALUES ($1,$2,$3,$4,$5,$6)
    `
 _, err = tx.ExecContext(ctx, query, flightID, seats, name, email, totalPrice, status)
 if err != nil {
  tx.Rollback()
  return fmt.Errorf("failed to insert booking: %w", err)
 }

 return tx.Commit()
}

func (r *BookingRepository) CancelBooking(ctx context.Context, bookingID int64) error {
 tx, err := r.db.BeginTx(ctx, nil)
 if err != nil {
  return err
 }

 query := DELETE FROM bookings WHERE id = $1
 res, err := tx.ExecContext(ctx, query, bookingID)
 if err != nil {
  tx.Rollback()
  return fmt.Errorf("failed to delete booking: %w", err)
 }

 rows, err := res.RowsAffected()
 if err != nil {
  tx.Rollback()
  return err
 }

 if rows == 0 {
  tx.Rollback()
  return fmt.Errorf("booking not found")
 }

 return tx.Commit()
}

func (r *BookingRepository) GetAllBookings(ctx context.Context) ([]Booking, error) {
 rows, err := r.db.QueryContext(ctx, "SELECT id, flight_id, seats, passenger_name, passenger_email, total_price, status FROM bookings")
 if err != nil {
  return nil, err
 }
 defer rows.Close()

 bookings := []Booking{}
 for rows.Next() {
  var b Booking
  if err := rows.Scan(&b.ID, &b.FlightID, &b.Seats, &b.PassengerName, &b.PassengerEmail, &b.TotalPrice, &b.Status); err != nil {
   return nil, err
  }
  bookings = append(bookings, b)
 }

 return bookings, nil
}

func (r *BookingRepository) GetBookingByID(ctx context.Context, id int64) (*Booking, error) {
 row := r.db.QueryRowContext(ctx, "SELECT id, flight_id, seats, passenger_name, passenger_email, total_price, status FROM bookings WHERE id=$1", id)
 var b Booking
 if err := row.Scan(&b.ID, &b.FlightID, &b.Seats, &b.PassengerName, &b.PassengerEmail, &b.TotalPrice, &b.Status); err != nil {
  if err == sql.ErrNoRows {
   return nil, nil
  }
  return nil, err
 }
 return &b, nil
}