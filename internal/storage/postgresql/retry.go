package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type retryClass int

const (
	nonRetriable retryClass = iota
	retriable
)

func classifyPGRead(err error) retryClass {
	if err == nil {
		return nonRetriable
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.CannotConnectNow:
			return retriable
		case pgerrcode.TransactionRollback,
			pgerrcode.SerializationFailure,
			pgerrcode.DeadlockDetected:
			return retriable
		}
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nonRetriable
	}
	return nonRetriable
}

func classifyPGWrite(err error) retryClass {
	if err == nil {
		return nonRetriable
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.TransactionRollback,
			pgerrcode.SerializationFailure,
			pgerrcode.DeadlockDetected:
			return retriable
		}
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nonRetriable
	}
	return nonRetriable
}

func withRetryRead(ctx context.Context, fn func() error) error {
	return retry.Do(
		fn,
		retry.Context(ctx),
		retry.Attempts(5),
		retry.Delay(200*time.Millisecond),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.RetryIf(func(err error) bool { return classifyPGRead(err) == retriable }),
	)
}

func withRetryWrite(ctx context.Context, fn func() error) error {
	return retry.Do(
		fn,
		retry.Context(ctx),
		retry.Attempts(5),
		retry.Delay(200*time.Millisecond),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.RetryIf(func(err error) bool { return classifyPGWrite(err) == retriable }),
	)
}
