package rollupcontract

import "errors"

var (
	ErrBatchAlreadyFinalized = errors.New("batch already finalized")
	ErrBatchAlreadyCommitted = errors.New("batch already committed")
)
