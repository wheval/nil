package concurrent

import (
	"context"
	"errors"
	"log"
	"time"
)

type workerState int8

const (
	_ workerState = iota
	workerStateRunning
	workerStatePaused
)

var ErrWorkerStopped = errors.New("worker was stopped")

type stateChangeRequest struct {
	newState workerState
	response chan bool
}

// Suspendable provides a mechanism for suspending and resuming periodic execution of an action.
type Suspendable struct {
	action   func(context.Context)
	interval time.Duration
	stateCh  chan stateChangeRequest
	stopped  chan struct{}
}

func NewSuspendable(action func(context.Context), interval time.Duration) *Suspendable {
	return &Suspendable{
		action:   action,
		interval: interval,
		stateCh:  make(chan stateChangeRequest),
		stopped:  make(chan struct{}),
	}
}

// Run executes a suspendable action periodically based on the provided interval until the context is canceled.
// It listens for pause and resume signals, halting and resuming execution accordingly.
func (s *Suspendable) Run(ctx context.Context, started chan<- struct{}) error {
	defer close(s.stopped)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	if started != nil {
		close(started)
	}

	state := workerStateRunning

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			s.action(ctx)

		case req := <-s.stateCh:
			s.onStateChange(ticker, &state, req)
		}
	}
}

func (s *Suspendable) onStateChange(ticker *time.Ticker, currentState *workerState, req stateChangeRequest) {
	defer close(req.response)

	switch {
	case req.newState == *currentState:
		// state remains unchanged, push false to the caller of Pause() / Resume()
		req.response <- false
		return

	case req.newState == workerStatePaused:
		ticker.Stop()

	case req.newState == workerStateRunning:
		ticker.Reset(s.interval)

	default:
		log.Panicf("unknown worker state: %d", req.newState)
	}

	*currentState = req.newState
	req.response <- true
}

func (s *Suspendable) Pause(ctx context.Context) (paused bool, err error) {
	return s.pushAndWait(ctx, workerStatePaused)
}

func (s *Suspendable) Resume(ctx context.Context) (resumed bool, err error) {
	return s.pushAndWait(ctx, workerStateRunning)
}

func (s *Suspendable) pushAndWait(ctx context.Context, newState workerState) (bool, error) {
	request := stateChangeRequest{newState: newState, response: make(chan bool)}

	select {
	case <-ctx.Done():
		return false, ctx.Err()

	case s.stateCh <- request:
		select {
		case <-ctx.Done():
			return false, ctx.Err()

		case stateWasChanged := <-request.response:
			return stateWasChanged, nil
		}

	case <-s.stopped:
		return false, ErrWorkerStopped
	}
}
