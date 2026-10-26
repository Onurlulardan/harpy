package runtimev1

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWireValuesSatisfyPublishedSchemas(t *testing.T) {
	errorEnvelope, err := NewErrorEnvelope("ROUTE_NOT_FOUND", "Route not found", "request-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	livez, err := NewLivez(ServiceCoreAPI)
	if err != nil {
		t.Fatal(err)
	}
	readyz, err := NewReadyz(ServiceCoreAPI, ReadinessReasonNone, ReadinessStates{
		Database: ReadinessStatusReady, DatabaseRole: ReadinessStatusReady,
		DeploymentPolicy: ReadinessStatusReady, SchemaContract: ReadinessStatusReady,
	})
	if err != nil {
		t.Fatal(err)
	}
	version, err := NewVersion(
		ServiceCoreAPI,
		"1.2.3",
		strings.Repeat("a", 40),
		"2026-07-13T00:00:00Z",
		EnvironmentProduction,
	)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		schema string
		value  any
	}{
		{"error-envelope.schema.json", errorEnvelope},
		{"livez.schema.json", livez},
		{"readyz.schema.json", readyz},
		{"version.schema.json", version},
	}
	for _, test := range tests {
		t.Run(test.schema, func(t *testing.T) {
			wire, err := json.Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}
			var decoded any
			if err := json.Unmarshal(wire, &decoded); err != nil {
				t.Fatal(err)
			}
			if err := compileSchema(t, test.schema).Validate(decoded); err != nil {
				t.Fatalf("wire value does not satisfy schema: %v", err)
			}
		})
	}
}
