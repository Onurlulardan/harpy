package notificationhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type successEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

func (handler *Handler) writeSuccess(w http.ResponseWriter, r *http.Request,
	status int, data any,
) {
	if err := writeJSON(w, status, handler.correlation(r), successEnvelope{true, data}); err != nil {
		handler.report(errors.New("notification HTTP response write failed"))
	}
}

func writeJSON(w http.ResponseWriter, status int, correlation string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal notification HTTP response: %w", err)
	}
	header := w.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")
	header.Set("Cache-Control", "no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Correlation-ID", correlation)
	header.Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	if _, err = w.Write(body); err != nil {
		return fmt.Errorf("write notification HTTP response: %w", err)
	}
	return nil
}
