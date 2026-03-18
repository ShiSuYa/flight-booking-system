CREATE TABLE bookings (
    id BIGSERIAL PRIMARY KEY,
    passenger_name VARCHAR NOT NULL,
    passenger_email VARCHAR NOT NULL,
    flight_id BIGINT NOT NULL,
    seats INT NOT NULL CHECK (seats > 0),
    total_price DECIMAL NOT NULL CHECK (total_price > 0),
    status VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);