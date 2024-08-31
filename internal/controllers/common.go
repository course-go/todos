package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/course-go/todos/internal/health"
)

func NotFound(w http.ResponseWriter, _ *http.Request) {
	code := http.StatusNotFound
	w.WriteHeader(code)
	w.Write(responseErrorBytes(code))
}

func (a API) Health(w http.ResponseWriter, _ *http.Request) {
	report := a.registry.Report()
	reportBytes, err := json.Marshal(report)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if report.Health == health.ERROR {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Write(reportBytes)
}
