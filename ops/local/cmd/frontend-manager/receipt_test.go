package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestReceiptRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "receipt.json")
	want := processReceipt{
		Version:         1,
		PID:             42,
		ProcessIdentity: "identity",
		FrontendRoot:    "frontend",
		ReadyURL:        "http://localhost:5173/",
		StartedAt:       time.Unix(123, 0).UTC(),
	}
	if err := writeReceipt(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := readReceipt(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("receipt = %#v, want %#v", got, want)
	}
}

func TestValidateReceiptAcceptsOnlyCanonicalAndLegacyLoopbackURLs(t *testing.T) {
	cfg := runtimeConfig{
		FrontendRoot: "frontend", ReadyURL: localFrontendReadyURL,
	}
	for _, readyURL := range []string{
		localFrontendReadyURL, legacyLocalFrontendReadyURL,
	} {
		receipt := processReceipt{FrontendRoot: "frontend", ReadyURL: readyURL}
		if err := validateReceipt(cfg, receipt); err != nil {
			t.Fatalf("ready URL %q was rejected: %v", readyURL, err)
		}
	}
	for _, receipt := range []processReceipt{
		{FrontendRoot: "other", ReadyURL: localFrontendReadyURL},
		{FrontendRoot: "frontend", ReadyURL: "http://example.test:5173/"},
	} {
		if err := validateReceipt(cfg, receipt); err == nil {
			t.Fatalf("foreign receipt was accepted: %#v", receipt)
		}
	}
}
