package telemetry

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

const serviceNamespace = "harpy"

func newResource(info buildinfo.Info, instanceID string) (*resource.Resource, error) {
	version := info.Version()
	if err := version.Validate(); err != nil {
		return nil, fmt.Errorf("telemetry build provenance is invalid")
	}
	if !validUUIDv4(instanceID) {
		return nil, fmt.Errorf("telemetry service instance ID is invalid")
	}
	attributes := []attribute.KeyValue{
		semconv.ServiceName(string(version.Service)),
		semconv.ServiceNamespace(serviceNamespace),
		semconv.ServiceVersion(version.Version),
		semconv.ServiceInstanceID(instanceID),
		semconv.DeploymentEnvironmentName(string(version.Environment)),
		semconv.VCSRefHeadRevision(version.Commit),
		attribute.Int("harpy.runtime.contract.version", version.ContractVersion),
	}
	return resource.NewWithAttributes(semconv.SchemaURL, attributes...), nil
}

func validUUIDv4(value string) bool {
	if len(value) != 36 || value[14] != '4' || value[19] < '8' || value[19] > 'b' {
		return false
	}
	for index := 0; index < len(value); index++ {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			if value[index] != '-' {
				return false
			}
			continue
		}
		if !lowerHex(value[index]) {
			return false
		}
	}
	return true
}

func lowerHex(value byte) bool {
	return value >= '0' && value <= '9' || value >= 'a' && value <= 'f'
}
