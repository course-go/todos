package metrics

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type Metrics struct {
	ProcessedRequests metric.Int64Counter
	RequestDuration   metric.Int64Histogram
}

func New(provider *sdkmetric.MeterProvider) (metrics *Metrics, err error) {
	meter := provider.Meter("todos.http")
	processedRequest, err := meter.Int64Counter("request.total")
	if err != nil {
		return nil, fmt.Errorf("failed creating total requests counter metric: %w", err)
	}

	requestDuration, err := meter.Int64Histogram("request.duration.ms")
	if err != nil {
		return nil, fmt.Errorf("failed creating request duration histogram: %w", err)
	}

	metrics = &Metrics{
		ProcessedRequests: processedRequest,
		RequestDuration:   requestDuration,
	}
	return metrics, nil
}
