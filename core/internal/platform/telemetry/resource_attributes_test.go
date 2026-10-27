package telemetry

import (
	"testing"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

func TestResourceUsesExactRuntimeIdentity(t *testing.T) {
	const instanceID = "00000000-0000-4000-8000-000000000000"
	resolved, err := newResource(telemetryBuildInfo(t), instanceID)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.SchemaURL() != semconv.SchemaURL {
		t.Fatalf("unexpected schema URL: %s", resolved.SchemaURL())
	}
	attributes := resourceStrings(resolved.Set().ToSlice())
	expected := map[string]string{
		"service.name":                "core-api",
		"service.namespace":           serviceNamespace,
		"service.version":             "0.0.0-dev",
		"service.instance.id":         instanceID,
		"deployment.environment.name": "test",
		"vcs.ref.head.revision":       "unknown",
	}
	for key, value := range expected {
		if attributes[key] != value {
			t.Fatalf("attribute %s: got %q, want %q", key, attributes[key], value)
		}
	}
}

func TestResourceRejectsInvalidInstanceIdentity(t *testing.T) {
	for _, instanceID := range []string{
		"", "00000000-0000-3000-8000-000000000000",
		"00000000-0000-4000-c000-000000000000",
		"00000000-0000-4000-8000-00000000000G",
	} {
		t.Run(instanceID, func(t *testing.T) {
			if _, err := newResource(telemetryBuildInfo(t), instanceID); err == nil {
				t.Fatalf("invalid instance ID accepted: %q", instanceID)
			}
		})
	}
}

func resourceStrings(attributes []attribute.KeyValue) map[string]string {
	result := make(map[string]string, len(attributes))
	for _, keyValue := range attributes {
		if keyValue.Value.Type() == attribute.STRING {
			result[string(keyValue.Key)] = keyValue.Value.AsString()
		}
	}
	return result
}
