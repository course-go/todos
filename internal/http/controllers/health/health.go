package health

import (
	"encoding/json"
	"net/http"

	"github.com/course-go/todos/internal/health"
)

type Controller struct {
	registry *health.Registry
}

func NewController(registry *health.Registry) *Controller {
	return &Controller{
		registry: registry,
	}
}

func (c *Controller) GetHealthController(w http.ResponseWriter, _ *http.Request) {
	report := c.registry.Report()

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
