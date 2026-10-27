package provider

import (
	"testing"
	"time"
)

func TestNullableProviderValuesRejectTypeDrift(t *testing.T) {
	var text nullableString
	if err := text.Scan("processing"); err != nil || !text.valid || text.value != "processing" {
		t.Fatalf("nullable string=%+v error=%v", text, err)
	}
	if err := text.Scan(nil); err != nil || text.valid {
		t.Fatalf("nullable string null=%+v error=%v", text, err)
	}
	if err := text.Scan(7); err == nil {
		t.Fatal("nullable string accepted type drift")
	}
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	var instant nullableTime
	if err := instant.Scan(now); err != nil || !instant.valid || instant.value != now {
		t.Fatalf("nullable time=%+v error=%v", instant, err)
	}
	if err := instant.Scan("2026-07-17T12:00:00Z"); err == nil {
		t.Fatal("nullable time accepted type drift")
	}
}
