package domain

import (
	"testing"
	"time"
)

func TestNormalizeObservationRejectsRawAndUnstableRequestData(t *testing.T) {
	now := time.Date(2026, 8, 1, 10, 0, 0, 123456000, time.UTC)
	base := Observation{ID: idFixture(7, 1), EnvironmentID: idFixture(4, 2),
		TenantID: idFixture(4, 3), ActorID: idFixture(4, 4), SourceIdentity: digestFixture(1),
		Method: "get", RoutePattern: "/api/forms/{form_id}", StatusCode: 200,
		DurationMicros: 1234, OccurredAt: now}
	normalized, err := NormalizeObservation(base)
	if err != nil || normalized.Method != "GET" || normalized.OccurredAt != now {
		t.Fatalf("normalized = %#v, %v", normalized, err)
	}
	first := normalized.Digest()
	second := normalized.Digest()
	if first != second || first == ([32]byte{}) {
		t.Fatal("fact digest is not deterministic")
	}
	for _, route := range []string{
		"/api/forms/123?token=secret", "/api/forms/123#fragment", "/api/forms/\nsecret",
	} {
		invalid := base
		invalid.RoutePattern = route
		if _, err = NormalizeObservation(invalid); err == nil {
			t.Fatalf("raw route %q was accepted", route)
		}
	}
}

func TestNormalizeObservationRequiresV7AndMicrosecondTime(t *testing.T) {
	value := Observation{ID: idFixture(4, 1), EnvironmentID: idFixture(4, 2),
		TenantID: idFixture(4, 3), ActorID: idFixture(4, 4),
		SourceIdentity: digestFixture(2), Method: "GET",
		RoutePattern: "/api/forms", StatusCode: 200, OccurredAt: time.Now().UTC()}
	if _, err := NormalizeObservation(value); err == nil {
		t.Fatal("non-v7 identity was accepted")
	}
	value.ID = idFixture(7, 5)
	value.OccurredAt = time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC)
	if _, err := NormalizeObservation(value); err == nil {
		t.Fatal("sub-microsecond timestamp was accepted")
	}
}

func idFixture(version, suffix byte) ID {
	var value ID
	value[6], value[8], value[15] = version<<4, 0x80, suffix
	return value
}

func digestFixture(value byte) [32]byte {
	var result [32]byte
	result[0] = value
	return result
}
