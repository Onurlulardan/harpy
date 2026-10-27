package api

import (
	"encoding/json"
	"testing"
)

func TestDirectionCanonicalSet(t *testing.T) {
	for name, want := range map[string]Direction{
		"asc": DirectionAscending, "desc": DirectionDescending,
	} {
		actual, found := ParseDirection(name)
		if !found || actual != want || actual.String() != name {
			t.Fatalf("direction %q did not round trip", name)
		}
	}
	if value, found := ParseDirection("ascending"); found || value != 0 {
		t.Fatal("unknown direction accepted")
	}
}

func TestCountModeCanonicalSet(t *testing.T) {
	for name, want := range map[string]CountMode{
		"none": CountModeNone, "exact": CountModeExact,
	} {
		actual, found := ParseCountMode(name)
		if !found || actual != want || actual.String() != name {
			t.Fatalf("count mode %q did not round trip", name)
		}
	}
	if value, found := ParseCountMode("estimated"); found || value != 0 {
		t.Fatal("unknown count mode accepted")
	}
}

func TestRequestEnumsCannotMarshalDirectly(t *testing.T) {
	for _, value := range []any{DirectionAscending, CountModeNone} {
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		}
	}
}
