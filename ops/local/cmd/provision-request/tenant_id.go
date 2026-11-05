package main

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var errInvalidTenant = errors.New("tenant provisioning request is invalid")

func parseTenantID(output []byte) (string, error) {
	text := strings.TrimSuffix(string(output), "\n")
	text = strings.TrimSuffix(text, "\r")
	if text == "" || strings.ContainsAny(text, "\r\n \t") {
		return "", errInvalidTenant
	}
	parsed, err := uuid.Parse(text)
	if err != nil || parsed == uuid.Nil || parsed.String() != text {
		return "", errInvalidTenant
	}
	return text, nil
}
