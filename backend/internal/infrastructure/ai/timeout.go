package ai

import (
	"context"
	"time"
)

const (
	DialTimeout       = 10 * time.Second
	TLSTimeout        = 5 * time.Second
	ResponseHeaderTimeout = 30 * time.Second
	RequestTimeout    = 60 * time.Second
	GenerationTimeout = 55 * time.Second
)

func ApplyTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
