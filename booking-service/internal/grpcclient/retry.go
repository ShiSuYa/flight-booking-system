package grpcclient

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Retry выполняет gRPC вызов с retry и exponential backoff
func Retry(ctx context.Context, operation func() error) error {

	backoffs := []time.Duration{
		0,
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
	}

	var err error

	for attempt := 0; attempt < len(backoffs); attempt++ {

		if attempt > 0 {
			time.Sleep(backoffs[attempt])
		}

		err = operation()

		if err == nil {
			return nil
		}

		st, ok := status.FromError(err)
		if !ok {
			return err
		}

		switch st.Code() {

		case codes.Unavailable, codes.DeadlineExceeded:
			// retry
			continue

		default:
			// другие ошибки не ретраим
			return err
		}
	}

	return err
}