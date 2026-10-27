package platformops

import (
	"net/http"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type healthComponentView struct {
	Name        string                      `json:"name"`
	URL         string                      `json:"url,omitempty"`
	Status      platformopsapi.HealthStatus `json:"status"`
	ResponseMS  int64                       `json:"response_ms"`
	LastChecked time.Time                   `json:"last_checked"`
	Error       string                      `json:"error,omitempty"`
}

type systemHealthView struct {
	Status         platformopsapi.HealthStatus `json:"status"`
	Timestamp      time.Time                   `json:"timestamp"`
	Services       []healthComponentView       `json:"services"`
	Infrastructure []healthComponentView       `json:"infrastructure"`
	HealthyCount   int                         `json:"healthy_count"`
	TotalCount     int                         `json:"total_count"`
}

func (handler *HTTPHandler) getSystemHealth(w http.ResponseWriter, r *http.Request) {
	environment, err := handler.environmentFor(r)
	if err == nil {
		_, err = handler.accessFor(r, environment, "platform.operations.health.read", "")
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	snapshot, err := handler.operations.ReadSystemHealth(r.Context())
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	view := systemHealthView{Status: snapshot.Status, Timestamp: snapshot.ObservedAt,
		Services:       healthComponentViews(snapshot.Services),
		Infrastructure: healthComponentViews(snapshot.Infrastructure)}
	view.TotalCount = len(view.Services) + len(view.Infrastructure)
	for _, item := range append(append([]healthComponentView{}, view.Services...), view.Infrastructure...) {
		if item.Status == platformopsapi.HealthHealthy {
			view.HealthyCount++
		}
	}
	handler.writeSuccess(w, r, http.StatusOK, view)
}

func healthComponentViews(values []platformopsapi.HealthComponent) []healthComponentView {
	result := make([]healthComponentView, len(values))
	for index, value := range values {
		result[index] = healthComponentView{Name: value.Name, URL: value.URL,
			Status: value.Status, ResponseMS: value.ResponseMillis,
			LastChecked: value.LastChecked, Error: value.Error}
	}
	return result
}
