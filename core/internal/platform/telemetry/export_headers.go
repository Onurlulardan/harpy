package telemetry

import (
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/baggage"
)

func parseExportHeaders(raw string) (map[string]string, error) {
	if raw == "" {
		return nil, nil
	}
	seen := make(map[string]bool)
	for _, segment := range strings.Split(raw, ",") {
		member := strings.TrimSpace(segment)
		key, _, found := strings.Cut(member, "=")
		canonical := http.CanonicalHeaderKey(strings.TrimSpace(key))
		if !found || canonical == "" {
			return nil, fmt.Errorf("OTLP headers contain an invalid member")
		}
		if seen[canonical] {
			return nil, fmt.Errorf("OTLP headers contain a duplicate name")
		}
		seen[canonical] = true
	}
	parsed, err := baggage.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("OTLP headers do not use the required encoding")
	}
	headers := make(map[string]string, parsed.Len())
	for _, member := range parsed.Members() {
		if len(member.Properties()) != 0 {
			return nil, fmt.Errorf("OTLP header metadata is not supported")
		}
		name := http.CanonicalHeaderKey(member.Key())
		if prohibitedExportHeader(name) {
			return nil, fmt.Errorf("OTLP headers contain a prohibited name")
		}
		headers[name] = member.Value()
	}
	return headers, nil
}

func prohibitedExportHeader(name string) bool {
	switch name {
	case "Baggage", "Connection", "Content-Length", "Content-Type", "Host",
		"Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer",
		"Traceparent", "Tracestate", "Transfer-Encoding", "Upgrade", "User-Agent":
		return true
	default:
		return false
	}
}
