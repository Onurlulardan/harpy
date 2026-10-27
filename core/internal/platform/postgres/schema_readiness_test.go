package postgres

import (
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestEvaluateSchemaContractFailsClosed(t *testing.T) {
	digest := strings.Repeat("a", 64)
	contract := runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion, Component: "core",
		RuntimeGeneration: 2, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256: digest, ExtensionManifestSHA256: digest,
		PostgresImageDigest: runtimev1.PostgresImageDigest,
	}
	validRun := "11111111-1111-4111-8111-111111111111"
	if reason := evaluateSchemaContract(
		contract, 1, 2, digest, digest, "ready", validRun, 1, true,
	); reason != SchemaReady {
		t.Fatalf("valid schema contract rejected: %s", reason)
	}
	tests := []struct {
		name                               string
		databaseGeneration, minimumRuntime int64
		migration, extension, status, run  string
		revision                           int64
		binaryReleaseKnown                 bool
		want                               SchemaReadinessReason
	}{
		{"fenced", 1, 2, digest, digest, "fenced", validRun, 1, true, SchemaFenced},
		{"database old", 0, 2, digest, digest, "ready", validRun, 1, true, SchemaContractMalformed},
		{"runtime old", 1, 3, digest, digest, "ready", validRun, 1, true, SchemaRuntimeTooOld},
		{"missing history", 1, 2, digest, digest, "ready", validRun, 1, false, SchemaManifestMismatch},
		{"malformed active hash", 1, 2, "bad", digest, "ready", validRun, 1, true, SchemaContractMalformed},
	}
	for _, test := range tests {
		reason := evaluateSchemaContract(
			contract, test.databaseGeneration, test.minimumRuntime, test.migration,
			test.extension, test.status, test.run, test.revision,
			test.binaryReleaseKnown,
		)
		if reason != test.want {
			t.Fatalf("%s reason = %s, want %s", test.name, reason, test.want)
		}
	}
}
