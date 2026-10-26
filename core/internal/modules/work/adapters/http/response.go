package workhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func (handler *Handler) complete(writer http.ResponseWriter, request *http.Request,
	spec routeSpec, status int, document []byte, err error,
) {
	if err != nil {
		handler.writeFailure(writer, request, err)
		return
	}
	if !json.Valid(document) {
		handler.writeFailure(writer, request, ErrUnavailable)
		return
	}
	body := document
	if !spec.compatibility {
		body = append([]byte(`{"success":true,"data":`), document...)
		body = append(body, '}')
	}
	if err = writeJSON(writer, status, handler.correlation(request), body); err != nil {
		handler.report(errors.New("Work HTTP response write failed"))
	}
}

func writeJSON(writer http.ResponseWriter, status int, correlation string, body []byte) error {
	if !json.Valid(body) {
		return fmt.Errorf("invalid Work HTTP response")
	}
	header := writer.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")
	header.Set("Cache-Control", "no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Correlation-ID", correlation)
	header.Set("Content-Length", strconv.Itoa(len(body)))
	writer.WriteHeader(status)
	if _, err := writer.Write(body); err != nil {
		return fmt.Errorf("write Work HTTP response: %w", err)
	}
	return nil
}
