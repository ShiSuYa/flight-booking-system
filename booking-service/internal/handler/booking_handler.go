package handler

import (
 "net/http"
 "strconv"

 "booking-service/internal/service"

 "github.com/gin-gonic/gin"
)

type BookingHandler struct {
 service *service.BookingService
 cb      *service.CircuitBreaker
}

func NewBookingHandler(s *service.BookingService) *BookingHandler {
 return &BookingHandler{
  service: s,
  cb:      service.NewCircuitBreaker(3, 5*time.Second),
 }
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
 type request struct {
  FlightID       int64   json:"flight_id"
  Seats          int32   json:"seats"
  PassengerName  string  json:"passenger_name"
  PassengerEmail string  json:"passenger_email"
  TotalPrice     float64 json:"total_price"
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
  req.TotalPrice,
  "CONFIRMED",
  h.cb,
 )
 if err != nil {
  c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
  return
 }

 c.JSON(http.StatusOK, gin.H{"status": "booking created"})
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
 type request struct {
  BookingID int64 json:"booking_id"
 }

 var req request
 if err := c.BindJSON(&req); err != nil {
  c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
  return
 }

 err := h.service.CancelBooking(c, req.BookingID)
 if err != nil {
  c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
  return
 }

 c.JSON(http.StatusOK, gin.H{"status": "booking cancelled"})
}

func (h *BookingHandler) GetAllBookings(c *gin.Context) {
 bookings, err := h.service.GetAllBookings(c)
 if err != nil {
  c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
  return
 }
 c.JSON(http.StatusOK, bookings)
}

func (h *BookingHandler) GetBookingByID(c *gin.Context) {
 idStr := c.Param("id")
 id, err := strconv.ParseInt(idStr, 10, 64)
 if err != nil {
  c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
  return
 }

 booking, err := h.service.GetBookingByID(c, id)
 if err != nil {
  c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
  return
 }
 if booking == nil {
  c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
  return
 }

 c.JSON(http.StatusOK, booking)
}