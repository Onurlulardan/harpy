package postgresv1

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestArtifactSizeLimitsAreExact(t *testing.T) {
	tests := []struct {
		name string
		got  int
		want int
	}{
		{"migration manifest", MaximumMigrationManifestBytes, 16 * 1024 * 1024},
		{"extension manifest", MaximumExtensionManifestBytes, 64 * 1024},
		{"roles manifest", MaximumRolesManifestBytes, 64 * 1024},
		{"privileges manifest", MaximumPrivilegesManifestBytes, 64 * 1024 * 1024},
		{"ownership manifest", MaximumOwnershipManifestBytes, 64 * 1024 * 1024},
		{"ownership chunk", MaximumOwnershipChunkBytes, 4 * 1024 * 1024},
		{"bootstrap operation", MaximumBootstrapOperationBytes, 64 * 1024},
		{"extension operation", MaximumExtensionOperationBytes, 64 * 1024},
		{"cleanup operation", MaximumCredentialCleanupOperationBytes, 64 * 1024},
	}
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("%s limit = %d, want %d", test.name, test.got, test.want)
		}
	}
}

func TestManifestLimitsCoverDeclaredCardinalityBudgets(t *testing.T) {
	entry := maximumSizedMigrationEntry()
	encodedEntry, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	const entryCount = 10000
	migrationUpperBound := 4096 + entryCount*(len(encodedEntry)+1)
	if migrationUpperBound > MaximumMigrationManifestBytes {
		t.Fatalf("migration schema maximum needs %d bytes", migrationUpperBound)
	}

	object := maximumSizedOwnershipObject()
	encodedObject, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	const inlineObjectCount = 4096
	ownershipUpperBound := 4096 + inlineObjectCount*(len(encodedObject)+1)
	if ownershipUpperBound > MaximumOwnershipManifestBytes {
		t.Fatalf("ownership inline maximum needs %d bytes", ownershipUpperBound)
	}

	const privilegeCapacity = 100000*560 + 10000*384 + 1024*1024
	if privilegeCapacity > MaximumPrivilegesManifestBytes {
		t.Fatalf("privilege cardinality budget needs %d bytes", privilegeCapacity)
	}
}

func maximumSizedMigrationEntry() MigrationEntry {
	return MigrationEntry{
		Owner: strings.Repeat("a", 64), Scope: "platform", Version: 99991231235959,
		TransactionMode: "non_transactional", Phase: "foundation",
		TargetDatabaseGeneration:     9007199254740991,
		NextMinimumRuntimeGeneration: 9007199254740991,
		Variants: MigrationVariants{
			Fresh:     &MigrationVariant{Path: maximumPath(".fresh.sql"), SHA256: strings.Repeat("f", 64)},
			V1Adopted: &MigrationVariant{Path: maximumPath(".v1_adopted.sql"), SHA256: strings.Repeat("f", 64)},
		},
	}
}

func maximumPath(suffix string) string {
	prefix := "99991231235959_"
	return prefix + strings.Repeat("a", 512-len(prefix)-len(suffix)) + suffix
}

func maximumSizedOwnershipObject() OwnershipObject {
	identifier := strings.Repeat("a", 63)
	arguments := make([]string, 64)
	for index := range arguments {
		arguments[index] = strings.Repeat("a", 128)
	}
	return OwnershipObject{
		Object: CatalogObject{
			Kind: "procedure", DatabaseName: identifier, SchemaName: &identifier,
			ObjectName: identifier, IdentityArguments: &arguments,
		},
		Owner: identifier, Classification: "first_party",
		Executor: "external_root_provisioner",
	}
}
