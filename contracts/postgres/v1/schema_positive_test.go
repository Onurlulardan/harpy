package postgresv1

import "testing"

func TestPostgresArtifactSchemasAcceptExactDocuments(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		value  map[string]any
	}{
		{"migration variants", "migration-manifest.schema.json", validMigrationManifest()},
		{"extensions", "extension-manifest.schema.json", validExtensionManifest()},
		{"fresh roles", "roles-manifest.schema.json", validRolesManifest("fresh")},
		{"adopted roles", "roles-manifest.schema.json", validRolesManifest("v1_adopted")},
		{"fresh privileges", "privileges-manifest.schema.json", validPrivilegesManifest("fresh")},
		{"adopted privileges", "privileges-manifest.schema.json", validPrivilegesManifest("v1_adopted")},
		{"inline ownership", "ownership-manifest.schema.json", validOwnershipManifest("fresh", "inline")},
		{"chunk root", "ownership-manifest.schema.json", validOwnershipManifest("v1_adopted", "chunked")},
		{"fresh chunk", "ownership-chunk.schema.json", validOwnershipChunk("fresh")},
		{"adopted chunk", "ownership-chunk.schema.json", validOwnershipChunk("v1_adopted")},
		{"fresh bootstrap", "bootstrap-operation.schema.json", validBootstrapOperation("fresh")},
		{"adopted bootstrap", "bootstrap-operation.schema.json", validBootstrapOperation("v1_adopted")},
		{"extension operation", "extension-operation.schema.json", validExtensionOperation()},
		{"cleanup operation", "credential-cleanup-operation.schema.json", validCleanupOperation()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := compileArtifactSchema(t, test.schema).Validate(test.value); err != nil {
				t.Fatalf("valid document rejected: %v", err)
			}
		})
	}
}

func TestPostgresArtifactSchemasRejectUnknownAndNull(t *testing.T) {
	documents := []struct {
		schema string
		value  map[string]any
	}{
		{"migration-manifest.schema.json", validMigrationManifest()},
		{"extension-manifest.schema.json", validExtensionManifest()},
		{"roles-manifest.schema.json", validRolesManifest("fresh")},
		{"privileges-manifest.schema.json", validPrivilegesManifest("fresh")},
		{"ownership-manifest.schema.json", validOwnershipManifest("fresh", "inline")},
		{"ownership-chunk.schema.json", validOwnershipChunk("fresh")},
		{"bootstrap-operation.schema.json", validBootstrapOperation("fresh")},
		{"extension-operation.schema.json", validExtensionOperation()},
		{"credential-cleanup-operation.schema.json", validCleanupOperation()},
	}
	for _, document := range documents {
		t.Run(document.schema+" unknown", func(t *testing.T) {
			value := cloneDocument(t, document.value)
			value["unexpected"] = true
			if err := compileArtifactSchema(t, document.schema).Validate(value); err == nil {
				t.Fatal("unknown field accepted")
			}
		})
		t.Run(document.schema+" null", func(t *testing.T) {
			value := cloneDocument(t, document.value)
			value["contract_version"] = nil
			if err := compileArtifactSchema(t, document.schema).Validate(value); err == nil {
				t.Fatal("null field accepted")
			}
		})
	}
}
