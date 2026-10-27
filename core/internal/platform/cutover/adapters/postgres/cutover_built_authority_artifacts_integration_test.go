//go:build integration

package postgres

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

const builtAuthorityOperationID = "32000000-0000-4000-8000-000000000010"

func buildBuiltAuthorityArtifacts(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	signer builtCutoverSigner, directory string, now time.Time,
) builtCutoverAuthorityArtifacts {
	t.Helper()
	bootstrap, bootstrapSignature := writeBuiltCutoverBootstrapAuthority(
		t, directory, signer, now,
	)
	operation := buildBuiltAuthorityOperation(
		t, ctx, database, bootstrap.digest, now,
	)
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
		t, directory, signer, now, bootstrap, operationArtifact,
	)
	return builtCutoverAuthorityArtifacts{
		releaseArtifact: release, releaseSignature: releaseSignature,
		bootstrapAuthority: bootstrap, bootstrapSignature: bootstrapSignature,
		operationArtifact: operationArtifact, operationSignature: operationSignature,
		receiptArtifact: builtCutoverArtifact{
			path: filepath.Join(directory, "authority-receipt.json"),
		},
		receiptSignature: builtCutoverArtifact{
			path: filepath.Join(directory, "authority-receipt.sig"),
		},
	}
}

func buildBuiltAuthorityOperation(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	bootstrapSHA256 string, now time.Time,
) cutoverv1.AuthorityAdoptionOperation {
	t.Helper()
	preview := previewBuiltAuthorityAdoption(
		t, ctx, database, builtAuthorityQuiescence{},
	)
	objects := make([]cutoverv1.AuthorityAdoptionObject, len(preview.preObjects))
	for index := range objects {
		identity, before, after := builtAuthorityObjectIdentities[index],
			preview.preObjects[index], preview.postObjects[index]
		objects[index] = cutoverv1.AuthorityAdoptionObject{
			Kind: identity.kind, SchemaName: identity.schema, ObjectName: identity.name,
			IdentityArguments: identity.arguments, Action: "adopt", IntroducedGeneration: 1,
			PreAdoptionOwner: before.owner, AdoptedOwner: after.owner,
			PreDefinitionSHA256: before.definition, AdoptedDefinitionSHA256: after.definition,
			PreAdoptionACLSHA256: before.acl, AdoptedACLSHA256: after.acl,
		}
	}
	dependencies := make([]cutoverv1.AuthorityCatalogDependency, len(preview.dependencies))
	for index, state := range preview.dependencies {
		identity := builtAuthorityDependencyIdentities[index]
		dependencies[index] = cutoverv1.AuthorityCatalogDependency{
			Kind: identity.kind, SchemaName: identity.schema, ObjectName: identity.name,
			IdentityArguments: identity.arguments, Owner: state.owner,
			DefinitionSHA256: state.definition, ACLSHA256: state.acl,
		}
	}
	target, _ := readBuiltCutoverDatabaseIdentity(t, ctx, database)
	operation := cutoverv1.AuthorityAdoptionOperation{
		ContractVersion: "cutover-authority-adoption-operation.v1",
		OperationID:     builtAuthorityOperationID, ReleaseID: integrationReleaseID,
		EnvironmentID: integrationEnvironment, Mode: "initial_adoption",
		CatalogGeneration: 1, PreCatalogContractSHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		IssuedAt:  now.Add(-time.Minute).UTC().Format(time.RFC3339),
		ExpiresAt: now.Add(8 * time.Minute).UTC().Format(time.RFC3339), Target: target,
		MigrationManifestSHA256:  digestMarker('6'),
		BootstrapAuthoritySHA256: bootstrapSHA256,
		Executor:                 "external_root_provisioner", ExecutorSessionUser: "postgres",
		OwnershipMode: "superuser_direct_per_object", MembershipPolicy: "always_zero",
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
		t.Fatalf("validate built authority operation: %v", err)
	}
	return operation
}

func builtAuthorityQuiescencePolicy(
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
