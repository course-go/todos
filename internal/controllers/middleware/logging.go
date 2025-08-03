package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.RequestURI
			method := r.Method
			next.ServeHTTP(w, r)

			duration := time.Since(start)
			logger.Info("handled HTTP request",
				"uri", uri,
				"method", method,
				"duration", duration,
			)
		})
	}
}
