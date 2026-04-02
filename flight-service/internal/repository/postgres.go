package repository

import (
 "database/sql"
 "fmt"
 "time"

 _ "github.com/lib/pq"
)

func NewPostgresDB() (*sql.DB, error) {
 connStr := "postgres://postgres:postgres@flight-db:5432/flightdb?sslmode=disable"

 var db *sql.DB
 var err error

 for i := 0; i < 10; i++ {
  db, err = sql.Open("postgres", connStr)
  if err != nil {
   fmt.Println("DB open error:", err)
   time.Sleep(2 * time.Second)
   continue
  }

  err = db.Ping()
  if err == nil {
   fmt.Println("Connected to flight-db")
   return db, nil
  }

  fmt.Println("Waiting for flight-db...", err)
  time.Sleep(2 * time.Second)
 }

 return nil, fmt.Errorf("could not connect to flight-db after retries: %w", err)
}