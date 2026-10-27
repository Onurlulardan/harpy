//go:build integration

package main

import (
	"context"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

const (
	workflowReleaseID     = "22222222-2222-4222-8222-222222222222"
	workflowEnvironmentID = "33333333-3333-4333-8333-333333333333"
	workflowAuthorityID   = "77777777-7777-4777-8777-777777777779"
)

type workflowIdentity struct {
	kind, schema, name, arguments string
}

var workflowAuthorityObjects = []workflowIdentity{
	{"schema", "cutover_authority", "cutover_authority", ""},
	{"table", "cutover_authority", "authority_adoption_inventory", ""},
	{"table", "cutover_authority", "finalization_authorizations", ""},
	{"table", "cutover_authority", "finalization_receipts", ""},
	{"function", "cutover_authority", "catalog_definition_sha256_v1", "text,oid"},
	{"function", "cutover_authority", "catalog_acl_sha256_v1", "text,oid"},
	{"function", "cutover_authority", "assert_adopted_catalog_v1", "text"},
	{"function", "cutover_authority", "finalization_state_sha256_v1", "uuid,uuid"},
	{"function", "cutover_authority", "record_credential_receipts_v1",
		"uuid,uuid,text,bytea,text,bytea"},
	{"function", "cutover_authority", "observe_finalizer_authority_v1", "uuid,uuid"},
	{"function", "cutover_authority", "assert_completed_evidence_v1",
		"uuid,uuid,bytea"},
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
}

func buildWorkflowAuthorityOperation(
	t *testing.T, ctx context.Context, database workflowDatabase,
	now time.Time, authoritySHA, migrationSHA string,
) cutoverv1.AuthorityAdoptionOperation {
	t.Helper()
	objects := make([]cutoverv1.AuthorityAdoptionObject, len(workflowAuthorityObjects))
	for index, identity := range workflowAuthorityObjects {
		objects[index] = cutoverv1.AuthorityAdoptionObject{
			Kind: identity.kind, SchemaName: identity.schema, ObjectName: identity.name,
			IdentityArguments: identity.arguments, Action: "adopt", IntroducedGeneration: 1,
			PreAdoptionOwner: "harpy_owner",
			AdoptedOwner:     "harpy_cutover_authority_owner",
		}
	}
	pre := readWorkflowObjectStates(t, ctx, database.admin, objects)
	transaction, err := database.admin.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := transferAdoptedObjects(ctx, transaction, objects); err != nil {
		t.Fatal(err)
	}
	post := readWorkflowObjectStates(t, ctx, transaction, objects)
	if err := transaction.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	for index := range objects {
		objects[index].PreDefinitionSHA256 = pre[index].definition
		objects[index].PreAdoptionACLSHA256 = pre[index].acl
		objects[index].AdoptedDefinitionSHA256 = post[index].definition
		objects[index].AdoptedACLSHA256 = post[index].acl
	}
	dependencies := workflowDependencies(t, ctx, database.admin)
	operation := cutoverv1.AuthorityAdoptionOperation{
		ContractVersion: "cutover-authority-adoption-operation.v1",
		OperationID:     workflowAuthorityID, ReleaseID: workflowReleaseID,
		EnvironmentID: workflowEnvironmentID, Mode: "initial_adoption",
		CatalogGeneration: 1,
		PreCatalogContractSHA256: "e3b0c44298fc1c149afbf4c8996fb924" +
			"27ae41e4649b934ca495991b7852b855",
		IssuedAt:  now.Add(-time.Minute).Format(time.RFC3339),
		ExpiresAt: now.Add(10 * time.Minute).Format(time.RFC3339),
		Target:    database.target, MigrationManifestSHA256: migrationSHA,
		BootstrapAuthoritySHA256: authoritySHA, Executor: "external_root_provisioner",
		ExecutorSessionUser: "postgres", OwnershipMode: "superuser_direct_per_object",
		MembershipPolicy: "always_zero", Objects: objects, Dependencies: dependencies,
		AuthorityOwnerRole: cutoverv1.AuthorityOwnerRoleState{
			ConnectionLimit: -1, PasswordState: "null", ValidUntil: "infinity",
			OwnedObjectCount: len(objects),
		},
	}
	operation.PostCatalogContractSHA256 =
		cutoverv1.AuthorityOperationCatalogContractSHA256(objects)
	operation.DependencyContractSHA256 =
		cutoverv1.AuthorityDependencyContractSHA256(dependencies)
	operation.QuiescencePolicy = workflowQuiescencePolicy(objects)
	if err := operation.Validate(); err != nil {
		t.Fatalf("validate workflow authority operation: %v", err)
	}
	return operation
}

func readWorkflowObjectStates(
	t *testing.T, ctx context.Context, query catalogQuerier,
	objects []cutoverv1.AuthorityAdoptionObject,
) []catalogState {
	t.Helper()
	states, err := readObjectStates(ctx, query, objects)
	if err != nil {
		t.Fatal(err)
	}
	return states
}

func workflowQuiescencePolicy(
	objects []cutoverv1.AuthorityAdoptionObject,
) cutoverv1.AuthorityQuiescencePolicy {
	pre := cutoverv1.AuthorityQuiescedRoleState{
		Login: true, ConnectionLimit: -1, PasswordState: "scram_sha_256",
		ValidUntil: "infinity", MembershipCount: 1,
	}
	temporary := pre
	temporary.Login, temporary.ConnectionLimit = false, 0
	return cutoverv1.AuthorityQuiescencePolicy{
		Roles: []cutoverv1.AuthorityQuiescedRolePolicy{{
			Principal: "harpy_migrator", PreState: pre, TemporaryState: temporary,
			PostState: pre, PasswordVerifierMustRemainUnchanged: true,
		}},
		TerminateSessions: true, RequireZeroRemainingSessions: true,
		RestoreBeforeCommit:         true,
		ExpectedPreNamespaceSHA256:  cutoverv1.AuthorityOperationPreNamespaceSHA256(objects),
		ExpectedPostNamespaceSHA256: cutoverv1.AuthorityOperationPostNamespaceSHA256(objects),
	}
}
