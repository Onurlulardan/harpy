package postgresv1

import "testing"

func TestPostgresArtifactSemanticValidatorsAcceptExactContracts(t *testing.T) {
	tests := []struct {
		name  string
		value semanticContract
	}{
		{"migration", validMigrationContract(t)},
		{"extension", validExtensionContract(t)},
		{"fresh roles", validRolesContract(t, "fresh")},
		{"adopted roles", validRolesContract(t, "v1_adopted")},
		{"fresh privileges", validPrivilegesContract(t, "fresh")},
		{"adopted privileges", validPrivilegesContract(t, "v1_adopted")},
		{"inline ownership", validOwnershipContract(t, "fresh", "inline")},
		{"chunked ownership", validOwnershipContract(t, "v1_adopted", "chunked")},
		{"ownership chunk", validOwnershipChunkContract(t, "fresh")},
		{"fresh bootstrap", validBootstrapContract(t, "fresh")},
		{"adopted bootstrap", validBootstrapContract(t, "v1_adopted")},
		{"extension operation", validExtensionOperationContract(t)},
		{"cleanup operation", validCleanupContract(t)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.value.Validate(); err != nil {
				t.Fatalf("valid semantic contract rejected: %v", err)
			}
		})
	}
}

func TestMigrationSelectionBindsOriginPathAndHash(t *testing.T) {
	manifest := validMigrationContract(t)
	selected, err := manifest.Select(DatabaseOriginV1Adopted)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || selected[0].Origin != DatabaseOriginV1Adopted ||
		selected[0].Path != manifest.Entries[0].Variants.V1Adopted.Path ||
		selected[0].SHA256 != manifest.Entries[0].Variants.V1Adopted.SHA256 {
		t.Fatal("selected migration does not preserve exact origin artifact binding")
	}
	selected[0].Path = "mutated"
	if manifest.Entries[0].Variants.V1Adopted.Path == "mutated" {
		t.Fatal("selected migration aliases manifest state")
	}
}
