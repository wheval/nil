//go:build race

package tests

import "time"

const (
	ReceiptWaitTimeout    = time.Minute
	ReceiptPollInterval   = time.Second
	BlockWaitTimeout      = 30 * time.Second
	BlockPollInterval     = time.Second
	ShardTickWaitTimeout  = 300 * time.Second
	ShardTickPollInterval = 10 * time.Second
)
