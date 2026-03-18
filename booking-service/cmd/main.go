package main

import (
    "log"
    "net/http"

    "booking-service/internal/handler"
    "booking-service/internal/repository"
    "booking-service/internal/service"

    "github.com/gin-gonic/gin"
)

func main() {
    // ----------------------------
    // Подключение к БД
    // ----------------------------
    db, err := repository.NewPostgresDB()
    if err != nil {
        log.Fatalf("failed to connect to DB: %v", err)
    }

    bookingRepo := repository.NewBookingRepository(db)
    bookingService := service.NewBookingService(bookingRepo)

    // ----------------------------
    // Handler (без gRPC и Circuit Breaker)
    // ----------------------------
    bookingHandler := handler.NewBookingHandler(bookingService)

    // ----------------------------
    // Роуты
    // ----------------------------
    router := gin.Default()

    router.POST("/bookings", bookingHandler.CreateBooking)
    router.POST("/bookings/cancel", bookingHandler.CancelBooking)

    log.Println("Booking Service started on port 8080")

    if err := http.ListenAndServe(":8080", router); err != nil {
        log.Fatalf("failed to start server: %v", err)
    }
}