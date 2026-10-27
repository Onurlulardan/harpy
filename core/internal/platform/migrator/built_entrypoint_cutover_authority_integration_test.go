//go:build integration

package migrator

import (
	"strings"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

var builtEntrypointAuthorityObjects = [][4]string{
	{"schema", "cutover_authority", "cutover_authority", ""},
	{"table", "cutover_authority", "authority_adoption_inventory", ""},
	{"table", "cutover_authority", "finalization_authorizations", ""},
	{"table", "cutover_authority", "finalization_receipts", ""},
	{"table", "cutover_authority", "full_rollout_completion_receipts", ""},
	{"function", "cutover_authority", "catalog_definition_sha256_v1", "text,oid"},
	{"function", "cutover_authority", "catalog_acl_sha256_v1", "text,oid"},
	{"function", "cutover_authority", "assert_adopted_catalog_v1", "text"},
	{"function", "cutover_authority", "finalization_state_sha256_v1", "uuid,uuid"},
	{"function", "cutover_authority", "record_credential_receipts_v1",
		"uuid,uuid,text,bytea,text,bytea"},
	{"function", "cutover_authority", "observe_finalizer_authority_v1", "uuid,uuid"},
	{"function", "cutover_authority", "assert_completed_evidence_v1", "uuid,uuid,bytea"},
	{"function", "cutover_authority", "assert_tenant_completion_v1", "uuid,uuid"},
	{"function", "cutover_authority", "assert_terminal_evidence_v1",
		"uuid,uuid,text,bytea"},
	{"function", "cutover_authority", "validate_finalization_envelope_v1",
		"uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea"},
	{"function", "cutover_authority", "authorize_finalization_v1",
		"uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea"},
	{"function", "cutover_authority", "consume_finalization_authorization_v1",
		"uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea"},
	{"function", "cutover_authority", "finalize_operation_v1",
		"uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea"},
	{"function", "cutover_authority", "verify_finalization_receipt_v1",
		"uuid,uuid,uuid,text,text"},
	{"function", "cutover_authority", "record_full_rollout_completion_v1",
		"uuid,uuid,uuid,text,text,text,text,text,text,text,text,text,text," +
			"timestampwithtimezone,timestampwithtimezone,timestampwithtimezone"},
}

var builtEntrypointAuthorityDependencies = [][4]string{
	{"schema", "platform", "platform", ""},
	{"table", "platform", "cutover_operations", ""},
	{"table", "platform", "cutover_operation_relations", ""},
	{"table", "platform", "cutover_operation_checkpoint_keys", ""},
	{"table", "platform", "cutover_operation_columns", ""},
	{"table", "platform", "cutover_operation_procedures", ""},
	{"table", "platform", "cutover_initial_copy_receipts", ""},
	{"table", "platform", "cutover_batch_receipts", ""},
	{"table", "platform", "cutover_copy_checkpoints", ""},
	{"table", "platform", "applied_change_receipts", ""},
	{"table", "platform", "cutover_change_journal", ""},
	{"table", "platform", "cutover_credential_receipts", ""},
	{"table", "platform", "tenant_migration_requirement_catalogs", ""},
	{"table", "platform", "tenant_migration_requirements", ""},
	{"table", "platform", "tenant_migration_requirement_targets", ""},
	{"table", "platform", "tenant_migration_checkpoints", ""},
	{"table", "platform", "tenants", ""},
	{"table", "platform", "release_generations", ""},
	{"table", "platform", "schema_contracts", ""},
}

func writeBuiltEntrypointCutoverAuthorityOperation(
	t *testing.T, directory string, now time.Time,
	catalog builtEntrypointArtifactCatalog, authority builtEntrypointArtifact,
	signer builtEntrypointSigner,
) (builtEntrypointArtifact, builtEntrypointArtifact) {
	t.Helper()
	operation := builtEntrypointCutoverAuthorityOperation(now, catalog, authority.digest)
	document := writeBuiltEntrypointArtifact(
		t, directory, "cutover-authority-operation.json", operation,
	)
	signature, err := signer.signer.SignAt(document.document, now.Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	signed := writeBuiltEntrypointArtifact(
		t, directory, "cutover-authority-operation.sig", cutoverv1.AuthorityAdoptionSignature{
			ContractVersion: "cutover-authority-adoption-operation-signature.v1",
			ManifestSHA256:  document.digest, Signature: cutoverv1.Signature{
				Algorithm: signature.Algorithm, KeyID: signature.KeyID, Value: signature.Value,
			},
		},
	)
	return document, signed
}

func builtEntrypointCutoverAuthorityOperation(
	now time.Time, catalog builtEntrypointArtifactCatalog, authoritySHA256 string,
) cutoverv1.AuthorityAdoptionOperation {
	digest := strings.Repeat("a", 64)
	objects := make([]cutoverv1.AuthorityAdoptionObject, len(builtEntrypointAuthorityObjects))
	for index, identity := range builtEntrypointAuthorityObjects {
		objects[index] = cutoverv1.AuthorityAdoptionObject{
			Kind: identity[0], SchemaName: identity[1], ObjectName: identity[2],
			IdentityArguments: identity[3], Action: "adopt", IntroducedGeneration: 1,
			PreAdoptionOwner:    "harpy_owner",
			AdoptedOwner:        "harpy_cutover_authority_owner",
			PreDefinitionSHA256: digest, AdoptedDefinitionSHA256: digest,
			PreAdoptionACLSHA256: digest, AdoptedACLSHA256: digest,
		}
	}
	dependencies := make([]cutoverv1.AuthorityCatalogDependency,
		len(builtEntrypointAuthorityDependencies))
	for index, identity := range builtEntrypointAuthorityDependencies {
		dependencies[index] = cutoverv1.AuthorityCatalogDependency{
			Kind: identity[0], SchemaName: identity[1], ObjectName: identity[2],
			IdentityArguments: identity[3], Owner: "harpy_owner",
			DefinitionSHA256: digest, ACLSHA256: digest,
		}
	}
	operation := cutoverv1.AuthorityAdoptionOperation{
		ContractVersion: "cutover-authority-adoption-operation.v1",
		OperationID:     "77777777-7777-4777-8777-777777777779",
		ReleaseID:       builtEntrypointReleaseID, EnvironmentID: builtEntrypointEnvironmentID,
		Mode: "initial_adoption", CatalogGeneration: 1,
		PreCatalogContractSHA256: "e3b0c44298fc1c149afbf4c8996fb924" +
			"27ae41e4649b934ca495991b7852b855",
		IssuedAt:  now.Add(-time.Minute).Format(time.RFC3339),
		ExpiresAt: now.Add(9 * time.Minute).Format(time.RFC3339),
		Target: cutoverv1.DatabaseIdentity{DatabaseName: builtEntrypointDatabaseName,
			DatabaseOrigin: "fresh", SystemIdentifier: "1", DatabaseOID: 1},
		MigrationManifestSHA256:  catalog.migration.digest,
		BootstrapAuthoritySHA256: authoritySHA256, Executor: "external_root_provisioner",
		ExecutorSessionUser: "postgres", OwnershipMode: "superuser_direct_per_object",
		MembershipPolicy: "always_zero", AuthorityOwnerRole: cutoverv1.AuthorityOwnerRoleState{
			ConnectionLimit: -1, PasswordState: "null", ValidUntil: "infinity",
			OwnedObjectCount: len(objects),
		},
		Objects: objects, Dependencies: dependencies,
	}
	operation.PostCatalogContractSHA256 =
		cutoverv1.AuthorityOperationCatalogContractSHA256(objects)
	operation.DependencyContractSHA256 =
		cutoverv1.AuthorityDependencyContractSHA256(dependencies)
	operation.QuiescencePolicy = builtEntrypointAuthorityQuiescence(objects)
	return operation
}

func TestBuiltEntrypointCutoverAuthorityOperationValidates(t *testing.T) {
	digest := strings.Repeat("a", 64)
	catalog := builtEntrypointArtifactCatalog{
		migration: builtEntrypointArtifact{digest: digest},
	}
	operation := builtEntrypointCutoverAuthorityOperation(
		time.Now().UTC(), catalog, digest,
	)
	if err := operation.Validate(); err != nil {
		t.Fatalf("validate built cutover authority operation: %v", err)
	}
}

func builtEntrypointAuthorityQuiescence(
	objects []cutoverv1.AuthorityAdoptionObject,
) cutoverv1.AuthorityQuiescencePolicy {
	atRest := cutoverv1.AuthorityQuiescedRoleState{
		Login: true, ConnectionLimit: -1, PasswordState: "scram_sha_256",
		ValidUntil: "infinity", MembershipCount: 1,
	}
	temporary := atRest
	temporary.Login, temporary.ConnectionLimit = false, 0
	return cutoverv1.AuthorityQuiescencePolicy{
		Roles: []cutoverv1.AuthorityQuiescedRolePolicy{{
			Principal: "harpy_migrator", PreState: atRest,
			TemporaryState: temporary, PostState: atRest,
			PasswordVerifierMustRemainUnchanged: true,
		}},
		TerminateSessions: true, RequireZeroRemainingSessions: true,
		RestoreBeforeCommit:         true,
		ExpectedPreNamespaceSHA256:  cutoverv1.AuthorityOperationPreNamespaceSHA256(objects),
		ExpectedPostNamespaceSHA256: cutoverv1.AuthorityOperationPostNamespaceSHA256(objects),
	}
}
