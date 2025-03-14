package telemetry

import (
	"context"
	"os"

	"github.com/NilFoundation/nil/nil/internal/telemetry/internal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type (
	Config = internal.Config

	Meter = metric.Meter
)

func NewDefaultConfig() *Config {
	// https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_service_name
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = os.Args[0]
	}
	return &Config{
		ServiceName: serviceName,
	}
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
