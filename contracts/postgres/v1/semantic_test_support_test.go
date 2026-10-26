package postgresv1

import (
	"encoding/json"
	"testing"
)

type semanticContract interface {
	Validate() error
}

func typedFixture[T any](t *testing.T, value map[string]any) T {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var target T
	if err := json.Unmarshal(encoded, &target); err != nil {
		t.Fatal(err)
	}
	return target
}

func validMigrationContract(t *testing.T) MigrationManifest {
	return typedFixture[MigrationManifest](t, validMigrationManifest())
}

func validExtensionContract(t *testing.T) ExtensionManifest {
	return typedFixture[ExtensionManifest](t, validExtensionManifest())
}

func validRolesContract(t *testing.T, origin string) RolesManifest {
	return typedFixture[RolesManifest](t, validRolesManifest(origin))
}

func validPrivilegesContract(t *testing.T, origin string) PrivilegesManifest {
	return typedFixture[PrivilegesManifest](t, validPrivilegesManifest(origin))
}

func validOwnershipContract(t *testing.T, origin, format string) OwnershipManifest {
	return typedFixture[OwnershipManifest](t, validOwnershipManifest(origin, format))
}

func validOwnershipChunkContract(t *testing.T, origin string) OwnershipChunk {
	return typedFixture[OwnershipChunk](t, validOwnershipChunk(origin))
}

func validBootstrapContract(t *testing.T, origin string) BootstrapOperation {
	return typedFixture[BootstrapOperation](t, validBootstrapOperation(origin))
}

func validExtensionOperationContract(t *testing.T) ExtensionOperation {
	return typedFixture[ExtensionOperation](t, validExtensionOperation())
}

func validCleanupContract(t *testing.T) CredentialCleanupOperation {
	return typedFixture[CredentialCleanupOperation](t, validCleanupOperation())
}
