package testaide

import (
	"context"
	"fmt"
	"time"
)

func WaitFor(ctx context.Context, signal <-chan struct{}, timeout time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-signal:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout after %s", timeout)
	}
}
