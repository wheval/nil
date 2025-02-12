package core

import (
	"time"

	"github.com/NilFoundation/nil/nil/internal/telemetry"
)

const (
	DefaultTaskRpcEndpoint = "tcp://127.0.0.1:8530"
)

type Config struct {
	RpcEndpoint             string
	TaskListenerRpcEndpoint string
	PollingDelay            time.Duration
	ProposerParams          *ProposerParams
	Telemetry               *telemetry.Config
}

func NewDefaultConfig() *Config {
	return &Config{
		RpcEndpoint:             "tcp://127.0.0.1:8529",
		TaskListenerRpcEndpoint: DefaultTaskRpcEndpoint,
		PollingDelay:            time.Second,
		ProposerParams:          NewDefaultProposerParams(),
		Telemetry: &telemetry.Config{
			ServiceName: "sync_committee",
		},
	}
}
