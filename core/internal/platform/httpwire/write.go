package httpwire

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

func WriteJSON(
	w http.ResponseWriter,
	status int,
	correlationID requestmeta.CorrelationID,
	payload any,
	headOnly bool,
) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal HTTP response: %w", err)
	}
	header := w.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")
	header.Set("Cache-Control", "no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set(requestmeta.CorrelationHeader, correlationID.String())
	header.Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	if headOnly {
		return nil
	}
	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("write HTTP response: %w", err)
	}
	return nil
}
