package concurrent

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunWithTimeout_AllFunctionsSucceed(t *testing.T) {
	t.Parallel()

	err := RunWithTimeout(
		t.Context(),
		1*time.Second,
		WithSource(func(ctx context.Context) error {
			return nil
		}),
		WithSource(func(ctx context.Context) error {
			return nil
		}))
	require.NoError(t, err)
}

func TestRunWithTimeout_FunctionFails(t *testing.T) {
	t.Parallel()

	err := RunWithTimeout(
		t.Context(),
		1*time.Second,
		WithSource(func(ctx context.Context) error {
			return errors.New("failure")
		}),
		WithSource(func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return nil
		}))
	require.Error(t, err)
	errorLines := strings.Split(err.Error(), "\n")
	require.Contains(t, errorLines[0], "failure")
	require.Contains(t, errorLines[2], "nil/nil/common/concurrent/utils_test.go:34")
}

func TestRunWithTimeout_TimeoutExceededWithContextCheck(t *testing.T) {
	t.Parallel()

	err := RunWithTimeout(
		t.Context(),
		10*time.Millisecond,
		WithSource(func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return nil
		}))
	require.Error(t, err)
	require.Contains(t, err.Error(), "context deadline exceeded")
}

func TestRunWithTimeout_TimeoutExceededWithoutContextCheck(t *testing.T) {
	t.Parallel()

	err := RunWithTimeout(
		t.Context(),
		10*time.Millisecond,
		WithSource(func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		}))
	require.NoError(t, err)
}
