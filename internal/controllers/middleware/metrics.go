package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/course-go/todos/internal/controllers/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func Metrics(metrics *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.RequestURI
			method := r.Method
			next.ServeHTTP(w, r)

			duration := time.Since(start)

			ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), time.Second)
			defer cancel()

			set := attribute.NewSet(
				attribute.KeyValue{
					Key:   "method",
					Value: attribute.StringValue(method),
				},
				attribute.KeyValue{
					Key:   "endpoint",
					Value: attribute.StringValue(uri),
				},
			)
			attributes := metric.WithAttributeSet(set)
			metrics.ProcessedRequests.Add(ctx, 1, attributes)
			metrics.RequestDuration.Record(ctx, duration.Milliseconds(), attributes)
		})
	}
}
