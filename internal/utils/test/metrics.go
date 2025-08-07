package test

import (
	"testing"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

func NewMetricProvider(t *testing.T) *metric.MeterProvider {
	t.Helper()

	exporter, err := prometheus.New()
	if err != nil {
		t.Fatalf("failed creating prometheus exporter: %v", err)
	}

	return metric.NewMeterProvider(metric.WithReader(exporter))
}
