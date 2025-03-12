package internal

type Config struct {
	ServiceName string `yaml:"serviceName,omitempty"`

	ExportMetrics bool   `yaml:"exportMetrics,omitempty"`
	GrpcEndpoint  string `yaml:"grpcEndpoint,omitempty"`

	PrometheusPort int `yaml:"prometheusPort,omitempty"`
}
