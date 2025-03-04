package core

import (
	"github.com/NilFoundation/nil/nil/internal/telemetry"
)

const (
	DefaultTaskRpcEndpoint = "tcp://127.0.0.1:8530"
)

type Config struct {
	RpcEndpoint             string
	TaskListenerRpcEndpoint string
	AggregatorConfig        AggregatorConfig
	ProposerParams          ProposerParams
	Telemetry               *telemetry.Config
}

func NewDefaultConfig() *Config {
	return &Config{
		RpcEndpoint:             "tcp://127.0.0.1:8529",
		TaskListenerRpcEndpoint: DefaultTaskRpcEndpoint,
		AggregatorConfig:        NewDefaultAggregatorConfig(),
		ProposerParams:          NewDefaultProposerParams(),
		Telemetry: &telemetry.Config{
			ServiceName: "sync_committee",
		},
	}
}
