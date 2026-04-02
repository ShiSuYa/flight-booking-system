package main

import (
 "log"
 "net/http"
 "time"

 "booking-service/internal/handler"
 "booking-service/internal/repository"
 "booking-service/internal/service"
 "booking-service/internal/grpcclient"

 "github.com/gin-gonic/gin"
)

func main() {
 db, err := repository.NewPostgresDB()
 if err != nil {
  log.Fatalf("failed to connect to DB: %v", err)
 }


 bookingRepo := repository.NewBookingRepository(db)

 flightClient := grpcclient.NewFlightClient()

 bookingService := service.NewBookingService(bookingRepo, flightClient)


 bookingHandler := handler.NewBookingHandler(bookingService)

 router := gin.Default()
 router.POST("/bookings", bookingHandler.CreateBooking)
 router.POST("/bookings/cancel", bookingHandler.CancelBooking)
 router.GET("/bookings", bookingHandler.GetAllBookings)
 router.GET("/bookings/:id", bookingHandler.GetBookingByID)

 srv := &http.Server{
  Addr:         ":8080",
  Handler:      router,
  ReadTimeout:  10 * time.Second,
  WriteTimeout: 10 * time.Second,
 }

 log.Println("Booking Service started on port 8080")
 if err := srv.ListenAndServe(); err != nil {
  log.Fatalf("failed to start server: %v", err)
 }
}