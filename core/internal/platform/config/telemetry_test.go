package config

import (
	"strings"
	"testing"
)

func TestTelemetryRequiresSecureProductionEndpoint(t *testing.T) {
	production := withEnvironment(validAPIEnvironment(),
		"HARPY_ENVIRONMENT=production",
		"HARPY_POSTGRES_DSN=postgresql://user@example.invalid/harpy?sslmode=verify-full",
	)
	if _, err := loadAPI(production); err == nil || !strings.Contains(err.Error(), "OTLP_ENDPOINT") {
		t.Fatalf("expected required endpoint error, got %v", err)
	}
	if _, err := loadAPI(withEnvironment(production,
		"HARPY_OTLP_ENDPOINT=http://collector.invalid:4318",
		"HARPY_OTLP_INSECURE=true")); err == nil {
		t.Fatal("expected production HTTP endpoint rejection")
	}
	if _, err := loadAPI(withEnvironment(production,
		"HARPY_OTLP_ENDPOINT=https://collector.invalid:4318")); err != nil {
		t.Fatalf("secure production endpoint rejected: %v", err)
	}
}

func TestTelemetryEnforcesSchemeAndInsecurePair(t *testing.T) {
	tests := [][]string{
		{"HARPY_OTLP_ENDPOINT=http://collector.invalid:4318"},
		{"HARPY_OTLP_ENDPOINT=https://collector.invalid:4318", "HARPY_OTLP_INSECURE=true"},
		{"HARPY_OTLP_ENDPOINT=ftp://collector.invalid"},
		{"HARPY_OTLP_ENDPOINT=https://user@collector.invalid"},
		{"HARPY_OTLP_ENDPOINT=https://collector.invalid?token=x"},
	}
	for _, additions := range tests {
		if _, err := loadAPI(withEnvironment(validAPIEnvironment(), additions...)); err == nil {
			t.Fatalf("expected endpoint rejection for %#v", additions)
		}
	}
	if _, err := loadAPI(withEnvironment(validAPIEnvironment(),
		"HARPY_OTLP_ENDPOINT=http://collector.invalid:4318",
		"HARPY_OTLP_INSECURE=true")); err != nil {
		t.Fatalf("development HTTP endpoint rejected: %v", err)
	}
}

func TestTelemetryRejectsOptionsWithoutEndpoint(t *testing.T) {
	for _, option := range []string{
		"HARPY_OTLP_INSECURE=true",
		"HARPY_OTLP_HEADERS=authorization=value",
	} {
		if _, err := loadAPI(withEnvironment(validAPIEnvironment(), option)); err == nil {
			t.Fatalf("expected option %s to require endpoint", option)
		}
	}
}
