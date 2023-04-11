package lambda

import (
	"context"
	"errors"
	"time"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// CtxKey is how the request trace is store/retrieve in the context
const CtxKey ctxKey = 1

// RequestTrace is a unique value attach to each request
// RequestTrace contains information about the request and it's use in logging tracing
type RequestTrace struct {
	ID         string
	Now        time.Time
	StatusCode int
}

func GetRequestTrace(ctx context.Context) (*RequestTrace, error) {
	v, ok := ctx.Value(CtxKey).(*RequestTrace)
	if !ok {
		return nil, errors.New("lambda request trace not found")
	}

	return v, nil
}

func GetTraceID(ctx context.Context) string {
	v, ok := ctx.Value(CtxKey).(*RequestTrace)
	if !ok {
		return "00000000-0000-0000-0000-000000000000"
	}
	return v.ID
}
