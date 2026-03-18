FROM golang:1.26

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o flight-service ./cmd

EXPOSE 50051

CMD ["./flight-service"]