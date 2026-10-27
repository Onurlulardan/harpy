package runtimev1

import (
	"encoding/json"
	"testing"
)

func TestNewLivezMarshalsCanonicalWire(t *testing.T) {
	t.Parallel()
	livez, err := NewLivez(ServiceCoreAPI)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(livez)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"status":"alive","service":"core-api"}`
	if string(encoded) != want {
		t.Fatalf("unexpected wire: %s", encoded)
	}
}

func TestNewLivezRejectsUnknownService(t *testing.T) {
	t.Parallel()
	if _, err := NewLivez(Service("api")); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLivezMarshalFailsClosed(t *testing.T) {
	t.Parallel()
	livez := Livez{Status: LivenessStatus("up"), Service: ServiceCoreAPI}
	if _, err := json.Marshal(livez); err == nil {
		t.Fatal("invalid liveness crossed the JSON boundary")
	}
}
