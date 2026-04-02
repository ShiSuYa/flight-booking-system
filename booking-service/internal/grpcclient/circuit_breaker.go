package grpcclient

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

type State string

const (
	CLOSED    State = "CLOSED"
	OPEN      State = "OPEN"
	HALF_OPEN State = "HALF_OPEN"
)

type CircuitBreaker struct {
	mu sync.Mutex

	state State

	failCount int
	maxFails  int

	resetTimeout time.Duration
	lastFailTime time.Time
}

func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		state:        CLOSED,
		maxFails:     5,
		resetTimeout: 10 * time.Second,
	}
}

func (cb *CircuitBreaker) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {

		cb.mu.Lock()

		switch cb.state {

		case OPEN:
			if time.Since(cb.lastFailTime) > cb.resetTimeout {
				log.Println("CircuitBreaker: OPEN → HALF_OPEN")
				cb.state = HALF_OPEN
			} else {
				cb.mu.Unlock()
				return status.Error(codes.Unavailable, "Flight Service unavailable (circuit open)")
			}

		case HALF_OPEN:
			log.Println("CircuitBreaker: HALF_OPEN request")

		case CLOSED:
		}

		cb.mu.Unlock()

		err := invoker(ctx, method, req, reply, cc, opts...)

		cb.mu.Lock()
		defer cb.mu.Unlock()

		if err != nil {
			cb.failCount++
			cb.lastFailTime = time.Now()

			log.Printf("CircuitBreaker: error count = %d\n", cb.failCount)

			if cb.failCount >= cb.maxFails {
				log.Println("CircuitBreaker: CLOSED → OPEN")
				cb.state = OPEN
			}

			return err
		}

		if cb.state == HALF_OPEN {
			log.Println("CircuitBreaker: HALF_OPEN → CLOSED")
			cb.state = CLOSED
		}

		cb.failCount = 0
		return nil
	}
}
