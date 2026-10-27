package main

import (
	"testing"
	"time"
)

func TestParseConfigAcceptsRepeatableTargets(t *testing.T) {
	config, err := parseConfig([]string{
		"-http", "http://api:8000/livez",
		"-http", "https://worker:9000/readyz",
		"-tcp", "postgres:5432",
		"-ca-file", "/run/local/ca.crt",
		"-timeout", "45s",
	})
	if err != nil {
		t.Fatalf("parseConfig: %v", err)
	}
	if len(config.httpTargets) != 2 || len(config.tcpTargets) != 1 ||
		config.caFile != "/run/local/ca.crt" || config.timeout != 45*time.Second {
		t.Fatalf("unexpected config: %+v", config)
	}
}

func TestParseConfigRejectsUnsafeOrUnboundedInput(t *testing.T) {
	tests := [][]string{
		{},
		{"-http", "https://user:secret@api/readyz"},
		{"-http", "https://api/readyz?token=secret"},
		{"-http", "ftp://api/readyz"},
		{"-tcp", "postgres"},
		{"-tcp", "postgres:70000"},
		{"-tcp", "postgres:5432", "-timeout", "50ms"},
		{"-tcp", "postgres:5432", "unexpected"},
	}
	for _, arguments := range tests {
		if _, err := parseConfig(arguments); err == nil {
			t.Fatalf("invalid arguments accepted: %v", arguments)
		}
	}
}
