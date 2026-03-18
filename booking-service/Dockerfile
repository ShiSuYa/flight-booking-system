FROM golang:1.26

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o booking-service ./cmd

EXPOSE 8080

CMD ["./booking-service"]