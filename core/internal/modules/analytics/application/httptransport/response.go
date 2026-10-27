package httptransport

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

func (runtime *Runtime) JSON(
	writer http.ResponseWriter, request *http.Request, status int, data any,
) {
	if err := writeJSON(writer, status, runtime.Correlation(request),
		successEnvelope{Success: true, Data: data}); err != nil {
		runtime.Report(errors.New("analytics HTTP response write failed"))
	}
}

func (runtime *Runtime) PublicJSON(
	writer http.ResponseWriter, request *http.Request, body []byte,
) error {
	if len(body) == 0 || len(body) > 1<<20 || !json.Valid(body) {
		return fmt.Errorf("analytics public response is invalid")
	}
	header := writer.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")
	header.Set("Cache-Control", "no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Correlation-ID", runtime.Correlation(request))
	header.Set("Content-Length", strconv.Itoa(len(body)))
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(body); err != nil {
		return fmt.Errorf("write analytics public response: %w", err)
	}
	return nil
}

func writeJSON(writer http.ResponseWriter, status int, correlation string, value any) error {
	body, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal analytics HTTP response: %w", err)
	}
	header := writer.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")
	header.Set("Cache-Control", "no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Correlation-ID", correlation)
	header.Set("Content-Length", strconv.Itoa(len(body)))
	writer.WriteHeader(status)
	if _, err = writer.Write(body); err != nil {
		return fmt.Errorf("write analytics HTTP response: %w", err)
	}
	return nil
}
