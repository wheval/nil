package jsonrpc

import "errors"

// ErrTransactionDiscarded is returned when the transaction is discarded, along with the reason.
var ErrTransactionDiscarded = errors.New("transaction discarded")
