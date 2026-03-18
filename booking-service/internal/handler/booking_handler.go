package handler

import (
    "net/http"

    "booking-service/internal/service"

    "github.com/gin-gonic/gin"
)

type BookingHandler struct {
    service *service.BookingService
}

func NewBookingHandler(s *service.BookingService) *BookingHandler {
    return &BookingHandler{service: s}
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
    type request struct {
        FlightID       int64  `json:"flight_id"`
        Seats          int32  `json:"seats"`
        PassengerName  string `json:"passenger_name"`
        PassengerEmail string `json:"passenger_email"`
    }

    var req request
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    err := h.service.CreateBooking(
        c,
        req.FlightID,
        req.Seats,
        req.PassengerName,
        req.PassengerEmail,
    )

    if err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "error": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "booking created",
    })
}
