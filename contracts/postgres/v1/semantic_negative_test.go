package postgresv1

import "testing"

func TestMigrationSemanticValidationRejectsOrderAndSelectionDrift(t *testing.T) {
	manifest := validMigrationContract(t)
	second := manifest.Entries[0]
	second.Version--
	manifest.Entries = append(manifest.Entries, second)
	if err := manifest.Validate(); err == nil {
		t.Fatal("out-of-order migration accepted")
	}

	manifest = validMigrationContract(t)
	manifest.Entries[0].Variants.V1Adopted = nil
	if err := manifest.Validate(); err != nil {
		t.Fatalf("single-origin manifest should remain structurally valid: %v", err)
	}
	if _, err := manifest.Select(DatabaseOriginV1Adopted); err == nil {
		t.Fatal("missing selected origin variant accepted")
	}

	manifest = validMigrationContract(t)
	manifest.Entries[0].Variants.Fresh.Path =
		"core/internal/modules/tenancy/migrations/20260714010000_foundation.v1_adopted.sql"
	if err := manifest.Validate(); err == nil {
		t.Fatal("variant path with a different origin accepted")
	}
}

func TestInventorySemanticValidationRejectsDangerousDrift(t *testing.T) {
	extensions := validExtensionContract(t)
	extensions.Extensions = extensions.Extensions[1:]
	if err := extensions.Validate(); err == nil {
		t.Fatal("missing required pg_trgm extension accepted")
	}

	roles := validRolesContract(t, "fresh")
	roles.Roles[6].Superuser = false
	if err := roles.Validate(); err == nil {
		t.Fatal("role catalog drift accepted")
	}

	privileges := validPrivilegesContract(t, "fresh")
	privileges.Grants = privileges.Grants[:len(privileges.Grants)-1]
	if err := privileges.Validate(); err == nil {
		t.Fatal("missing direct login CONNECT grant accepted")
	}

	privileges = validPrivilegesContract(t, "fresh")
	create := privileges.Grants[0]
	create.Privileges = []string{"CREATE"}
	privileges.Grants = append([]ACLGrant{create}, privileges.Grants...)
	if err := privileges.Validate(); err == nil {
		t.Fatal("unordered ACL inventory accepted")
	}

	ownership := validOwnershipContract(t, "fresh", "inline")
	ownership.ObjectCount++
	if err := ownership.Validate(); err == nil {
		t.Fatal("ownership object count drift accepted")
	}
}

func TestOperationSemanticValidationRejectsTypedStepDrift(t *testing.T) {
	bootstrap := validBootstrapContract(t, "fresh")
	bootstrap.Steps[2].TransactionClass = "transactional_ddl"
	if err := bootstrap.Validate(); err == nil {
		t.Fatal("bootstrap transaction class drift accepted")
	}

	cleanup := validCleanupContract(t)
	cleanup.PrivilegedRoles[0], cleanup.PrivilegedRoles[1] =
		cleanup.PrivilegedRoles[1], cleanup.PrivilegedRoles[0]
	if err := cleanup.Validate(); err == nil {
		t.Fatal("cleanup privileged role order drift accepted")
	}
}
