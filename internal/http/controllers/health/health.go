package health

import (
	"encoding/json"
	"net/http"

	"github.com/course-go/todos/internal/health"
)

type HealthController struct {
	registry *health.Registry
}

func NewHealthController(registry *health.Registry) *HealthController {
	return &HealthController{
		registry: registry,
	}
}

func (hc *HealthController) GetHealthController(w http.ResponseWriter, _ *http.Request) {
	report := hc.registry.Report()

	reportBytes, err := json.Marshal(report)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if report.Health == health.ERROR {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	_, _ = w.Write(reportBytes)
}
