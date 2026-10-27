package application

import (
	"math"
	"regexp"
	"strings"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

var traceSpanIDPattern = regexp.MustCompile(`^[0-9a-f]{16,32}$`)

func validTraceObservation(value platformopsapi.TraceObservation) bool {
	if !validRecoveryUUID(value.EnvironmentID) ||
		!validCorrelationID(value.CorrelationID) ||
		!traceSpanIDPattern.MatchString(value.SpanID) ||
		(value.ParentSpanID != "" && !traceSpanIDPattern.MatchString(value.ParentSpanID)) ||
		len(value.ServiceName) < 1 || len(value.ServiceName) > 100 ||
		len(value.Operation) < 1 || len(value.Operation) > 256 ||
		len(value.Path) < 1 || len(value.Path) > 256 ||
		!strings.HasPrefix(value.Path, "/") || strings.ContainsAny(value.Path, "?#\\") ||
		value.StatusCode < 100 || value.StatusCode > 599 || len(value.Message) > 500 ||
		value.StartedAt.IsZero() || value.EndedAt.IsZero() || value.EndedAt.Before(value.StartedAt) ||
		value.DurationMillis < 0 || math.IsNaN(value.DurationMillis) ||
		math.IsInf(value.DurationMillis, 0) {
		return false
	}
	if !validTraceMethod(value.Method) || !validTraceLevel(value.Level) {
		return false
	}
	switch value.SpanKind {
	case "server", "client", "internal", "consumer", "producer":
		return true
	default:
		return false
	}
}

func validCorrelationID(value string) bool {
	if len(value) == 0 || len(value) > 64 || !asciiAlphaNumeric(value[0]) {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if !asciiAlphaNumeric(character) && character != '.' &&
			character != '_' && character != '-' {
			return false
		}
	}
	return true
}

func asciiAlphaNumeric(character byte) bool {
	return character >= 'A' && character <= 'Z' ||
		character >= 'a' && character <= 'z' ||
		character >= '0' && character <= '9'
}

func validTraceMethod(value string) bool {
	switch value {
	case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS":
		return true
	default:
		return false
	}
}

func validTraceLevel(value string) bool {
	switch value {
	case "debug", "info", "warn", "error", "fatal":
		return true
	default:
		return false
	}
}
