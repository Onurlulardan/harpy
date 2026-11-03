package postgres

import (
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestBaselineFoundationAdmitsAppliedGooseHistory(t *testing.T) {
	digest := strings.Repeat("b", 64)
	binary := runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion, Component: "core",
		RuntimeGeneration: 1, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256: digest, ExtensionManifestSHA256: digest,
		PostgresImageDigest: runtimev1.PostgresImageDigest,
	}
	result, err := assembleBaselineFoundation(
		"fresh", "7703580167595228103",
		"018f0000-0000-7000-8000-000000000102",
		"018f0000-0000-7000-8000-000000000101",
		runtimev1.VectorExtensionVersion, "platform_ext", "harpy_extension_admin",
		51019, 2, binary,
	)
	if err != nil {
		t.Fatal(err)
	}
	catalog := catalogPreflight{
		systemIdentifier: "7703580167595228103", databaseOID: 51019,
		vectorVersion: runtimev1.VectorExtensionVersion,
		vectorSchema:  "platform_ext", vectorOwner: "harpy_extension_admin",
	}
	origin, err := validateFoundationEvidence(result, catalog)
	if err != nil || origin != runtimev1.DatabaseOriginFresh {
		t.Fatalf("baseline evidence = %v, %v", origin, err)
	}
	if reason := foundationSchemaReadiness(result, binary); reason != SchemaReady {
		t.Fatalf("baseline readiness = %s", reason)
	}
	catalog.systemIdentifier = "1"
	if _, err := validateFoundationEvidence(result, catalog); err == nil {
		t.Fatal("baseline evidence accepted a different cluster")
	}
	if _, err := assembleBaselineFoundation(
		"fresh", "1", "018f0000-0000-7000-8000-000000000102",
		"018f0000-0000-7000-8000-000000000101",
		runtimev1.VectorExtensionVersion, "platform_ext", "harpy_extension_admin",
		1, 1, binary,
	); err == nil {
		t.Fatal("partial goose history was treated as a complete baseline")
	}
}
