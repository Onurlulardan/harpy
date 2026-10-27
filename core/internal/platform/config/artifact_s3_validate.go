package config

import (
	"net"
	"net/url"
	"strings"
)

func requiredArtifactText(values environmentValues, key string, maximum int) (string, error) {
	value, found := values.lookup(key)
	if !found || value == "" || len(value) > maximum || strings.TrimSpace(value) != value ||
		strings.ContainsAny(value, "\x00\r\n\t") {
		return "", configError(key, "is required and must be canonical")
	}
	return value, nil
}

func validArtifactEndpoint(value string, allowHTTP bool) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed == nil || parsed.Host == "" || parsed.User != nil ||
		parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" ||
		parsed.String() != value || parsed.Host != strings.ToLower(parsed.Host) {
		return false
	}
	return parsed.Scheme == "https" || allowHTTP && parsed.Scheme == "http"
}

func validArtifactRegion(value string) bool {
	if len(value) < 1 || len(value) > 63 || !lowercaseLetterOrDigit(value[0]) {
		return false
	}
	for index := 1; index < len(value); index++ {
		if !lowercaseLetterOrDigit(value[index]) && value[index] != '-' {
			return false
		}
	}
	return true
}

func validArtifactBucket(value string) bool {
	if len(value) < 3 || len(value) > 63 || net.ParseIP(value) != nil ||
		strings.Contains(value, "..") {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if label == "" || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if character != '-' && (character < 'a' || character > 'z') &&
				(character < '0' || character > '9') {
				return false
			}
		}
	}
	return true
}

func validArtifactCredential(value string) bool {
	if value == "" {
		return false
	}
	for index := 0; index < len(value); index++ {
		if value[index] < 0x21 || value[index] > 0x7e {
			return false
		}
	}
	return true
}
