package srv

import (
	"context"
	"time"
)

//go:generate bash ../scripts/generate_mock.sh Worker

type Worker interface {
	// Name returns the name of the Worker. This is typically used for logging and identification.
	Name() string

	// Run starts the worker, signaling its initialization through the started channel.
	Run(ctx context.Context, started chan<- struct{}) error
}

type WorkerLoop struct {
	name     string
	interval time.Duration
	action   func(ctx context.Context)
}

func NewWorkerLoop(name string, interval time.Duration, action func(ctx context.Context)) WorkerLoop {
	return WorkerLoop{
		name:     name,
		interval: interval,
		action:   action,
	}
}

func (w *WorkerLoop) Name() string {
	return w.name
}

func (w *WorkerLoop) Run(ctx context.Context, started chan<- struct{}) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	close(started)

	for {
		select {
		case <-ticker.C:
			w.action(ctx)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
