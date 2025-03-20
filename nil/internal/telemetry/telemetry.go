package telemetry

import (
	"context"

	"github.com/NilFoundation/nil/nil/internal/telemetry/internal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type (
	Config = internal.Config

	Meter = metric.Meter
)

func NewDefaultConfig() *Config {
	return &Config{}
}

func Init(ctx context.Context, config *Config) error {
	if config == nil {
		// no telemetry
		return nil
	}

	internal.StartPrometheusServer(config.PrometheusPort)

	return internal.InitMetrics(ctx, config)
}

func Shutdown(ctx context.Context) {
	internal.ShutdownMetrics(ctx)
}

func NewMeter(name string) Meter {
	return otel.Meter(name)
}
