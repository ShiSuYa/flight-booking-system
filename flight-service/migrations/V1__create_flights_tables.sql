CREATE TABLE flights (
    id BIGSERIAL PRIMARY KEY,
    flight_number VARCHAR NOT NULL,
    airline VARCHAR NOT NULL,
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_time TIMESTAMP NOT NULL,
    arrival_time TIMESTAMP NOT NULL,
    total_seats INT NOT NULL CHECK (total_seats > 0),
    available_seats INT NOT NULL CHECK (available_seats >= 0),
    price DECIMAL NOT NULL CHECK (price > 0),
    status VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE seat_reservations (
    id BIGSERIAL PRIMARY KEY,
    flight_id BIGINT NOT NULL,
    booking_id BIGINT UNIQUE NOT NULL,
    seats_reserved INT NOT NULL CHECK (seats_reserved > 0),
    status VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,

    CONSTRAINT fk_flight
        FOREIGN KEY (flight_id)
        REFERENCES flights(id)
);