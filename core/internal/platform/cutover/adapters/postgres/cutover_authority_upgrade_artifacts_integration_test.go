//go:build integration

package postgres

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

const builtAuthorityUpgradeOperationID = "32000000-0000-4000-8000-000000000011"

var builtAuthorityUpgradeIdentity = builtAuthorityIdentity{
	kind: "function", schema: "cutover_authority",
	name: "historical_receipt_probe_v2", arguments: "uuid",
}

func buildBuiltAuthorityUpgradeArtifacts(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	signer builtCutoverSigner, initial builtCutoverAuthorityArtifacts,
	directory string, now time.Time,
) builtCutoverAuthorityArtifacts {
	t.Helper()
	createBuiltAuthorityUpgradeProbe(t, ctx, database)
	operation := buildBuiltAuthorityUpgradeOperation(t, ctx, database, initial, now)
	operationArtifact := writeBuiltCutoverArtifact(
		t, directory, "authority-operation.json", operation,
	)
	operationSignature := writeBuiltCutoverArtifact(
		t, directory, "authority-operation.sig", cutoverv1.AuthorityAdoptionSignature{
			ContractVersion: "cutover-authority-adoption-operation-signature.v1",
			ManifestSHA256:  operationArtifact.digest,
			Signature: signBuiltCutoverDocument(
				t, signer, operationArtifact.document, now.Add(-time.Minute),
			),
		},
	)
	release, releaseSignature := writeBuiltCutoverRelease(
		t, directory, signer, now, initial.bootstrapAuthority, operationArtifact,
	)
	return builtCutoverAuthorityArtifacts{
		releaseArtifact: release, releaseSignature: releaseSignature,
		bootstrapAuthority: initial.bootstrapAuthority,
		bootstrapSignature: initial.bootstrapSignature,
		operationArtifact:  operationArtifact, operationSignature: operationSignature,
		receiptArtifact: builtCutoverArtifact{
			path: filepath.Join(directory, "authority-receipt.json"),
		},
		receiptSignature: builtCutoverArtifact{
			path: filepath.Join(directory, "authority-receipt.sig"),
		},
	}
}

func createBuiltAuthorityUpgradeProbe(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	_, err := database.admin.Exec(ctx, `CREATE FUNCTION
cutover_authority.historical_receipt_probe_v2(uuid) RETURNS boolean
LANGUAGE sql IMMUTABLE SET search_path TO pg_catalog AS 'SELECT $1 IS NOT NULL';
ALTER FUNCTION cutover_authority.historical_receipt_probe_v2(uuid)
OWNER TO harpy_cutover_authority_owner;
REVOKE ALL ON FUNCTION cutover_authority.historical_receipt_probe_v2(uuid) FROM PUBLIC;`)
	if err != nil {
		t.Fatalf("create authority generation-two probe: %v", err)
	}
}

func buildBuiltAuthorityUpgradeOperation(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	initial builtCutoverAuthorityArtifacts, now time.Time,
) cutoverv1.AuthorityAdoptionOperation {
	t.Helper()
	var previous cutoverv1.AuthorityAdoptionReceipt
	if err := attestation.DecodeStrict(initial.receiptArtifact.document, &previous); err != nil {
		t.Fatalf("decode generation-one authority receipt: %v", err)
	}
	identities := append([]builtAuthorityIdentity(nil), builtAuthorityObjectIdentities...)
	identities = append(identities, builtAuthorityUpgradeIdentity)
	states := readBuiltAuthorityCatalogStates(t, ctx, database.admin, identities)
	objects := make([]cutoverv1.AuthorityAdoptionObject, len(states))
	for index, state := range states {
		action, generation := "retain", 2
		if index < len(previous.Objects) {
			generation = previous.Objects[index].IntroducedGeneration
		} else {
			action = "add"
		}
		objects[index] = cutoverv1.AuthorityAdoptionObject{
			Kind: state.identity.kind, SchemaName: state.identity.schema,
			ObjectName: state.identity.name, IdentityArguments: state.identity.arguments,
			Action: action, IntroducedGeneration: generation,
			PreAdoptionOwner: state.owner, AdoptedOwner: state.owner,
			PreDefinitionSHA256: state.definition, AdoptedDefinitionSHA256: state.definition,
			PreAdoptionACLSHA256: state.acl, AdoptedACLSHA256: state.acl,
		}
	}
	dependencies := builtAuthorityUpgradeDependencies(t, ctx, database)
	target, _ := readBuiltCutoverDatabaseIdentity(t, ctx, database)
	previousSHA256, upgradeSHA256 := initial.receiptArtifact.digest, digestMarker('8')
	operation := cutoverv1.AuthorityAdoptionOperation{
		ContractVersion: "cutover-authority-adoption-operation.v1",
		OperationID:     builtAuthorityUpgradeOperationID, ReleaseID: integrationReleaseID,
		EnvironmentID: integrationEnvironment, Mode: "catalog_upgrade",
		PreviousCatalogGeneration: 1, CatalogGeneration: 2,
		PreCatalogContractSHA256:       previous.PostCatalogContractSHA256,
		PreviousAuthorityReceiptSHA256: &previousSHA256,
		UpgradeArtifactSHA256:          &upgradeSHA256,
		IssuedAt:                       now.Add(-time.Minute).UTC().Format(time.RFC3339),
		ExpiresAt:                      now.Add(8 * time.Minute).UTC().Format(time.RFC3339), Target: target,
		MigrationManifestSHA256:  previous.MigrationManifestSHA256,
		BootstrapAuthoritySHA256: initial.bootstrapAuthority.digest,
		Executor:                 "external_root_provisioner", ExecutorSessionUser: "postgres",
		OwnershipMode: "superuser_set_local_role_append_only", MembershipPolicy: "always_zero",
		AuthorityOwnerRole: cutoverv1.AuthorityOwnerRoleState{
			ConnectionLimit: -1, PasswordState: "null", ValidUntil: "infinity",
			OwnedObjectCount: len(objects),
		},
		Objects: objects, Dependencies: dependencies,
	}
	operation.PostCatalogContractSHA256 =
		cutoverv1.AuthorityOperationCatalogContractSHA256(objects)
	operation.DependencyContractSHA256 =
		cutoverv1.AuthorityDependencyContractSHA256(dependencies)
	operation.QuiescencePolicy = builtAuthorityQuiescencePolicy(objects)
	if err := operation.Validate(); err != nil {
		t.Fatalf("validate generation-two authority operation: %v", err)
	}
	return operation
}

func builtAuthorityUpgradeDependencies(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) []cutoverv1.AuthorityCatalogDependency {
	t.Helper()
	states := readBuiltAuthorityCatalogStates(
		t, ctx, database.admin, builtAuthorityDependencyIdentities,
	)
	dependencies := make([]cutoverv1.AuthorityCatalogDependency, len(states))
	for index, state := range states {
		dependencies[index] = cutoverv1.AuthorityCatalogDependency{
			Kind: state.identity.kind, SchemaName: state.identity.schema,
			ObjectName: state.identity.name, IdentityArguments: state.identity.arguments,
			Owner: state.owner, DefinitionSHA256: state.definition, ACLSHA256: state.acl,
		}
	}
	return dependencies
}
