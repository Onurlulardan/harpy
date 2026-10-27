package config

import (
	"math"
	"net"
	"strconv"
	"strings"
	"time"
)

func parseDuration(values environmentValues, key, fallback string, minimum, maximum time.Duration) (time.Duration, error) {
	raw := values.valueOr(key, fallback)
	if raw == "" || strings.TrimSpace(raw) != raw {
		return 0, configError(key, "must be a duration without surrounding whitespace")
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value < minimum || value > maximum {
		return 0, configError(key, "is outside the allowed duration range")
	}
	return value, nil
}

func parseInteger(values environmentValues, key, fallback string, minimum, maximum int64) (int64, error) {
	raw := values.valueOr(key, fallback)
	if !decimalDigits(raw) {
		return 0, configError(key, "must be an unsigned base-10 integer")
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < minimum || value > maximum {
		return 0, configError(key, "is outside the allowed integer range")
	}
	return value, nil
}

func parseRatio(values environmentValues, key, fallback string) (float64, error) {
	raw := values.valueOr(key, fallback)
	if !decimalRatio(raw) {
		return 0, configError(key, "must be a decimal ratio")
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
		return 0, configError(key, "is outside the allowed ratio range")
	}
	return value, nil
}

func parseBoolean(values environmentValues, key, fallback string) (bool, error) {
	raw := values.valueOr(key, fallback)
	switch raw {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, configError(key, "must be true or false")
	}
}

func parseListenAddress(values environmentValues, key, fallback string) (string, string, error) {
	raw := values.valueOr(key, fallback)
	if raw == "" || strings.TrimSpace(raw) != raw {
		return "", "", configError(key, "must be a host:port address")
	}
	host, port, err := net.SplitHostPort(raw)
	if err != nil || host == "" || strings.ContainsAny(host, " \t\r\n") || !decimalDigits(port) {
		return "", "", configError(key, "must be a host:port address")
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 || port != strconv.FormatUint(portNumber, 10) {
		return "", "", configError(key, "port must be between 1 and 65535")
	}
	return raw, strconv.FormatUint(portNumber, 10), nil
}

func decimalDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func decimalRatio(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		return false
	}
	for _, part := range parts {
		if !decimalDigits(part) {
			return false
		}
	}
	return true
}
