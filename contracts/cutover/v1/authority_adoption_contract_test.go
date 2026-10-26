package cutoverv1

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAuthorityAdoptionContractsAcceptExactBoundary(t *testing.T) {
	operation := validAuthorityAdoptionOperation()
	if err := operation.Validate(); err != nil {
		t.Fatalf("valid authority operation rejected: %v", err)
	}
	receipt := validAuthorityAdoptionReceipt(operation)
	if err := receipt.Validate(); err != nil {
		t.Fatalf("valid authority receipt rejected: %v", err)
	}
	assertAuthoritySchema(t, "authority-adoption-operation.schema.json", operation)
	assertAuthoritySchema(t, "authority-adoption-receipt.schema.json", receipt)
}

func TestAuthorityAdoptionContractsRejectMembershipAndObjectDrift(t *testing.T) {
	operation := validAuthorityAdoptionOperation()
	operation.MembershipPolicy = "temporary_set_role"
	if err := operation.Validate(); err == nil {
		t.Fatal("temporary authority-owner membership accepted")
	}
	operation = validAuthorityAdoptionOperation()
	operation.Objects[0].ObjectName = "platform"
	if err := operation.Validate(); err == nil {
		t.Fatal("authority object identity drift accepted")
	}
	receipt := validAuthorityAdoptionReceipt(validAuthorityAdoptionOperation())
	receipt.AuthorityOwnerMembershipCount = 1
	if err := receipt.Validate(); err == nil {
		t.Fatal("authority receipt with membership accepted")
	}
}

func validAuthorityAdoptionOperation() AuthorityAdoptionOperation {
	digest := strings.Repeat("a", 64)
	objects := make([]AuthorityAdoptionObject, len(authorityObjectOrder))
	for index, identity := range authorityObjectOrder {
		objects[index] = AuthorityAdoptionObject{
			Kind: identity[0], SchemaName: identity[1], ObjectName: identity[2],
			IdentityArguments: identity[3], Action: "adopt", IntroducedGeneration: 1,
			PreAdoptionOwner: "harpy_owner", AdoptedOwner: authorityOwner,
			PreDefinitionSHA256: digest, AdoptedDefinitionSHA256: digest,
			PreAdoptionACLSHA256: digest, AdoptedACLSHA256: digest,
		}
	}
	operation := AuthorityAdoptionOperation{
		ContractVersion: "cutover-authority-adoption-operation.v1",
		OperationID:     "10000000-0000-4000-8000-000000000021",
		ReleaseID:       "10000000-0000-4000-8000-000000000022",
		EnvironmentID:   "10000000-0000-4000-8000-000000000023",
		Mode:            "initial_adoption", PreviousCatalogGeneration: 0, CatalogGeneration: 1,
		PreCatalogContractSHA256: emptyCatalogSHA256, PostCatalogContractSHA256: digest,
		IssuedAt: "2026-07-14T10:00:00Z", ExpiresAt: "2026-07-14T10:15:00Z",
		Target: DatabaseIdentity{DatabaseName: "harpy", DatabaseOrigin: "fresh",
			SystemIdentifier: "123456789", DatabaseOID: 16384},
		MigrationManifestSHA256: digest, BootstrapAuthoritySHA256: digest,
		Executor: "external_root_provisioner", ExecutorSessionUser: "postgres",
		OwnershipMode: "superuser_direct_per_object", MembershipPolicy: "always_zero",
		AuthorityOwnerRole: validAuthorityOwnerRole(),
		Objects:            objects,
	}
	operation.PostCatalogContractSHA256 = AuthorityOperationCatalogContractSHA256(objects)
	operation.Dependencies = validAuthorityDependencies()
	operation.DependencyContractSHA256 = AuthorityDependencyContractSHA256(operation.Dependencies)
	operation.QuiescencePolicy = validAuthorityQuiescencePolicy(objects)
	return operation
}

func validAuthorityAdoptionReceipt(
	operation AuthorityAdoptionOperation,
) AuthorityAdoptionReceipt {
	digest := strings.Repeat("b", 64)
	objects := make([]AuthorityAdoptionReceiptObject, len(operation.Objects))
	for index, object := range operation.Objects {
		objects[index] = AuthorityAdoptionReceiptObject{
			Kind: object.Kind, SchemaName: object.SchemaName, ObjectName: object.ObjectName,
			IdentityArguments: object.IdentityArguments, IntroducedGeneration: 1,
			ObjectOID: uint32(17000 + index), Owner: authorityOwner,
			DefinitionSHA256: object.AdoptedDefinitionSHA256,
			ACLSHA256:        object.AdoptedACLSHA256,
		}
	}
	receipt := AuthorityAdoptionReceipt{
		ContractVersion: "cutover-authority-adoption-receipt.v1",
		ReceiptID:       "10000000-0000-4000-8000-000000000024",
		OperationID:     operation.OperationID, ReleaseID: operation.ReleaseID,
		EnvironmentID: operation.EnvironmentID, Target: operation.Target,
		Mode: operation.Mode, PreviousCatalogGeneration: operation.PreviousCatalogGeneration,
		CatalogGeneration:         operation.CatalogGeneration,
		PreCatalogContractSHA256:  operation.PreCatalogContractSHA256,
		PostCatalogContractSHA256: operation.PostCatalogContractSHA256,
		OperationSHA256:           digest, MigrationManifestSHA256: operation.MigrationManifestSHA256,
		ReleaseManifestSHA256: digest, BootstrapAuthoritySHA256: operation.BootstrapAuthoritySHA256,
		AdoptedAt: "2026-07-14T10:01:00Z", Result: "succeeded",
		ExecutorSessionUser: operation.ExecutorSessionUser, ExecutorSuperuser: true,
		PreAdoptionMembershipCount: 0, PreAdoptionOwnedObjectCount: 0,
		AuthorityOwnerMembershipCount: 0,
		AuthorityOwnerRole:            validAuthorityOwnerRole(),
		CatalogInstanceSHA256:         digest, Objects: objects,
	}
	receipt.PostCatalogContractSHA256 = AuthorityReceiptCatalogContractSHA256(objects)
	receipt.CatalogInstanceSHA256 = AuthorityReceiptCatalogInstanceSHA256(objects)
	receipt.Dependencies = make(
		[]AuthorityCatalogDependencyReceipt, len(operation.Dependencies),
	)
	for index, dependency := range operation.Dependencies {
		receipt.Dependencies[index] = AuthorityCatalogDependencyReceipt{
			AuthorityCatalogDependency: dependency, ObjectOID: uint32(18000 + index),
		}
	}
	receipt.DependencyContractSHA256 =
		AuthorityReceiptDependencyContractSHA256(receipt.Dependencies)
	receipt.DependencyInstanceSHA256 =
		AuthorityReceiptDependencyInstanceSHA256(receipt.Dependencies)
	receipt.QuiescenceEvidence = validAuthorityQuiescenceEvidence(operation, objects)
	return receipt
}

func validAuthorityDependencies() []AuthorityCatalogDependency {
	digest := strings.Repeat("c", 64)
	dependencies := make([]AuthorityCatalogDependency, len(authorityDependencyOrder))
	for index, identity := range authorityDependencyOrder {
		dependencies[index] = AuthorityCatalogDependency{
			Kind: identity[0], SchemaName: identity[1], ObjectName: identity[2],
			IdentityArguments: identity[3], Owner: "harpy_owner",
			DefinitionSHA256: digest, ACLSHA256: digest,
		}
	}
	return dependencies
}

func validAuthorityOwnerRole() AuthorityOwnerRoleState {
	return AuthorityOwnerRoleState{
		ConnectionLimit: -1, PasswordState: "null", ValidUntil: "infinity",
		OwnedObjectCount: len(authorityObjectOrder),
	}
}

func validAuthorityQuiescencePolicy(
	objects []AuthorityAdoptionObject,
) AuthorityQuiescencePolicy {
	atRest := authorityMigratorAtRestState("scram_sha_256")
	return AuthorityQuiescencePolicy{
		Roles: []AuthorityQuiescedRolePolicy{{
			Principal: authorityQuiescedPrincipal, PreState: atRest,
			TemporaryState: authorityMigratorTemporaryState(atRest),
			PostState:      atRest, PasswordVerifierMustRemainUnchanged: true,
		}},
		TerminateSessions: true, RequireZeroRemainingSessions: true, RestoreBeforeCommit: true,
		ExpectedPreNamespaceSHA256:  AuthorityOperationPreNamespaceSHA256(objects),
		ExpectedPostNamespaceSHA256: AuthorityOperationPostNamespaceSHA256(objects),
	}
}

func validAuthorityQuiescenceEvidence(
	operation AuthorityAdoptionOperation, objects []AuthorityAdoptionReceiptObject,
) AuthorityQuiescenceEvidence {
	policy := operation.QuiescencePolicy.Roles[0]
	return AuthorityQuiescenceEvidence{
		Roles: []AuthorityQuiescedRoleEvidence{{
			Principal: policy.Principal, ObservedPreState: policy.PreState,
			ObservedTemporaryState: policy.TemporaryState, ObservedPostState: policy.PostState,
			PasswordVerifierUnchanged: true,
		}},
		ObservedPreNamespaceSHA256:  operation.QuiescencePolicy.ExpectedPreNamespaceSHA256,
		ObservedPostNamespaceSHA256: AuthorityReceiptNamespaceSHA256(objects),
		RestoredBeforeCommit:        true,
	}
}

func assertAuthoritySchema(t *testing.T, path string, value any) {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var decoded any
	if err := json.Unmarshal(document, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := compileCutoverSchema(t, path).Validate(decoded); err != nil {
		t.Fatalf("%s rejected: %v", path, err)
	}
}
