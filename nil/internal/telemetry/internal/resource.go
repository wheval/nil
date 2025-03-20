package internal

import (
	"os"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0" // Be sure otelcol support it before update
)

func NewResource(config *Config) (*resource.Resource, error) {
	name := config.ServiceName
	if name == "" {
		// https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_service_name
		name = os.Getenv("OTEL_SERVICE_NAME")
		if name == "" {
			name = os.Args[0]
		}
	}

	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(name),
		),
	)
}
