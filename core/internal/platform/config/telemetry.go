package config

import (
	"net/url"
	"strings"
	"time"
)

const telemetryHeaderLimit = 8 * 1024

type Telemetry struct {
	endpoint          string
	headers           Secret
	headersConfigured bool
	insecure          bool
	exportTimeout     time.Duration
	traceSampleRatio  float64
}

func loadTelemetry(values environmentValues, environment Environment) (Telemetry, error) {
	endpoint, endpointSet := values.lookup("HARPY_OTLP_ENDPOINT")
	insecure, err := parseBoolean(values, "HARPY_OTLP_INSECURE", "false")
	if err != nil {
		return Telemetry{}, err
	}
	headers, err := loadSecret(values, "HARPY_OTLP_HEADERS", false, telemetryHeaderLimit)
	if err != nil {
		return Telemetry{}, err
	}
	_, directHeaders := values.lookup("HARPY_OTLP_HEADERS")
	_, fileHeaders := values.lookup("HARPY_OTLP_HEADERS_FILE")
	headersConfigured := directHeaders || fileHeaders
	if headersConfigured && containsControl(headers.Value()) {
		return Telemetry{}, configError("HARPY_OTLP_HEADERS", "contains a prohibited control character")
	}
	if !endpointSet {
		if environment.productionLike() {
			return Telemetry{}, configError("HARPY_OTLP_ENDPOINT", "is required in this environment")
		}
		if insecure || headersConfigured {
			return Telemetry{}, configError("HARPY_OTLP_ENDPOINT", "is required when exporter options are set")
		}
	} else if err := validateOTLPEndpoint(endpoint, insecure, environment); err != nil {
		return Telemetry{}, err
	}
	exportTimeout, err := parseDuration(values, "HARPY_OTLP_EXPORT_TIMEOUT", "5s", time.Second, 30*time.Second)
	if err != nil {
		return Telemetry{}, err
	}
	sampleRatio, err := parseRatio(values, "HARPY_TRACE_SAMPLE_RATIO", "0.1")
	if err != nil {
		return Telemetry{}, err
	}
	return Telemetry{
		endpoint: endpoint, headers: headers, headersConfigured: headersConfigured,
		insecure: insecure, exportTimeout: exportTimeout, traceSampleRatio: sampleRatio,
	}, nil
}

func validateOTLPEndpoint(raw string, insecure bool, environment Environment) error {
	if raw == "" || strings.TrimSpace(raw) != raw {
		return configError("HARPY_OTLP_ENDPOINT", "must be an absolute HTTP URL")
	}
	endpoint, err := url.ParseRequestURI(raw)
	if err != nil || endpoint.Host == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return configError("HARPY_OTLP_ENDPOINT", "must be an absolute URL without credentials, query, or fragment")
	}
	switch endpoint.Scheme {
	case "http":
		if !insecure || environment.productionLike() {
			return configError("HARPY_OTLP_ENDPOINT", "HTTP requires insecure mode outside staging and production")
		}
	case "https":
		if insecure {
			return configError("HARPY_OTLP_INSECURE", "must be false for HTTPS")
		}
	default:
		return configError("HARPY_OTLP_ENDPOINT", "scheme must be HTTP or HTTPS")
	}
	return nil
}

func containsControl(value string) bool {
	for _, character := range []byte(value) {
		if character < 0x20 || character == 0x7f {
			return true
		}
	}
	return false
}

func (t Telemetry) Enabled() bool                { return t.endpoint != "" }
func (t Telemetry) Endpoint() string             { return t.endpoint }
func (t Telemetry) Headers() (Secret, bool)      { return t.headers, t.headersConfigured }
func (t Telemetry) Insecure() bool               { return t.insecure }
func (t Telemetry) ExportTimeout() time.Duration { return t.exportTimeout }
func (t Telemetry) TraceSampleRatio() float64    { return t.traceSampleRatio }
