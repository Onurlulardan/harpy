package attestation

import (
	"fmt"
	"strings"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func authorityInitialReceiptFixture() VerifiedAuthorityAdoptionReceipt {
	objects := []cutoverv1.AuthorityAdoptionReceiptObject{{
		Kind: "table", SchemaName: "cutover_authority", ObjectName: "catalog_v1",
		IntroducedGeneration: 1, ObjectOID: 17001,
		Owner:            "harpy_cutover_authority_owner",
		DefinitionSHA256: strings.Repeat("1", 64), ACLSHA256: strings.Repeat("2", 64),
	}}
	receipt := cutoverv1.AuthorityAdoptionReceipt{
		ReceiptID: "receipt-1", EnvironmentID: "environment-1",
		CatalogGeneration: 1, AdoptedAt: "2026-07-14T10:01:00Z",
		Target: authorityChainTarget(), Objects: objects,
	}
	receipt.PostCatalogContractSHA256 = cutoverv1.AuthorityReceiptCatalogContractSHA256(objects)
	receipt.CatalogInstanceSHA256 = cutoverv1.AuthorityReceiptCatalogInstanceSHA256(objects)
	operationObjects := []cutoverv1.AuthorityAdoptionObject{retainedAuthorityObject(objects[0])}
	operation := cutoverv1.AuthorityAdoptionOperation{
		QuiescencePolicy: authorityQuiescencePolicyFixture(operationObjects),
	}
	receipt.QuiescenceEvidence = authorityQuiescenceEvidenceFixture(operation, objects)
	return newVerifiedAuthorityAdoptionReceipt(receipt, strings.Repeat("3", 64), []byte("one"))
}

func authorityChainUpgradeFixture(
	previous VerifiedAuthorityAdoptionReceipt, generation int,
) (VerifiedAuthorityAdoptionOperation, cutoverv1.AuthorityAdoptionReceipt,
	VerifiedRelease, VerifiedBootstrapAuthority) {
	prior := previous.Value()
	previousDigest, artifact := previous.ReceiptSHA256(), strings.Repeat("4", 64)
	operationObjects := make([]cutoverv1.AuthorityAdoptionObject, 0, len(prior.Objects)+1)
	for _, object := range prior.Objects {
		operationObjects = append(operationObjects, retainedAuthorityObject(object))
	}
	definition, acl := strings.Repeat("5", 64), strings.Repeat("6", 64)
	operationObjects = append(operationObjects, cutoverv1.AuthorityAdoptionObject{
		Kind: "function", SchemaName: "cutover_authority",
		ObjectName: fmt.Sprintf("catalog_probe_v%d", generation), IdentityArguments: "uuid",
		Action:               "add",
		IntroducedGeneration: generation,
		PreAdoptionOwner:     "harpy_cutover_authority_owner",
		AdoptedOwner:         "harpy_cutover_authority_owner",
		PreDefinitionSHA256:  definition, AdoptedDefinitionSHA256: definition,
		PreAdoptionACLSHA256: acl, AdoptedACLSHA256: acl,
	})
	operation := cutoverv1.AuthorityAdoptionOperation{
		OperationID: "operation", ReleaseID: "release", EnvironmentID: prior.EnvironmentID,
		Mode: "catalog_upgrade", PreviousCatalogGeneration: generation - 1,
		CatalogGeneration: generation, PreCatalogContractSHA256: prior.PostCatalogContractSHA256,
		PreviousAuthorityReceiptSHA256: &previousDigest, UpgradeArtifactSHA256: &artifact,
		IssuedAt: "2026-07-14T10:00:00Z", ExpiresAt: "2026-07-14T10:10:00Z",
		Target: prior.Target, MigrationManifestSHA256: strings.Repeat("7", 64),
		BootstrapAuthoritySHA256: strings.Repeat("8", 64), ExecutorSessionUser: "postgres",
		OwnershipMode: "superuser_set_local_role_append_only",
		Objects:       operationObjects,
	}
	operation.AuthorityOwnerRole.OwnedObjectCount = len(operationObjects)
	operation.PostCatalogContractSHA256 = cutoverv1.AuthorityOperationCatalogContractSHA256(operationObjects)
	operation.Dependencies = authorityChainDependencies()
	operation.DependencyContractSHA256 =
		cutoverv1.AuthorityDependencyContractSHA256(operation.Dependencies)
	operation.QuiescencePolicy = authorityQuiescencePolicyFixture(operationObjects)
	operationDigest := strings.Repeat(fmt.Sprintf("%x", generation), 64)[:64]
	verifiedOperation := newVerifiedAuthorityAdoptionOperation(operation, operationDigest, nil)
	receiptObjects := append([]cutoverv1.AuthorityAdoptionReceiptObject(nil), prior.Objects...)
	receiptObjects = append(receiptObjects, cutoverv1.AuthorityAdoptionReceiptObject{
		Kind: "function", SchemaName: "cutover_authority",
		ObjectName:           fmt.Sprintf("catalog_probe_v%d", generation),
		IdentityArguments:    "uuid",
		IntroducedGeneration: generation, ObjectOID: uint32(17000 + generation),
		Owner: "harpy_cutover_authority_owner", DefinitionSHA256: definition, ACLSHA256: acl,
	})
	releaseDigest := strings.Repeat("9", 64)
	receipt := authorityChainReceipt(
		operation, operationDigest, releaseDigest, prior, receiptObjects,
	)
	for index, dependency := range operation.Dependencies {
		receipt.Dependencies = append(
			receipt.Dependencies, cutoverv1.AuthorityCatalogDependencyReceipt{
				AuthorityCatalogDependency: dependency, ObjectOID: uint32(18000 + index),
			},
		)
	}
	receipt.DependencyContractSHA256 =
		cutoverv1.AuthorityReceiptDependencyContractSHA256(receipt.Dependencies)
	receipt.DependencyInstanceSHA256 =
		cutoverv1.AuthorityReceiptDependencyInstanceSHA256(receipt.Dependencies)
	release := newVerifiedRelease(ReleaseManifest{}, releaseDigest)
	authority := VerifiedBootstrapAuthority{
		manifest:       runtimev1.BootstrapAuthorityManifest{},
		manifestDigest: operation.BootstrapAuthoritySHA256, verified: true,
	}
	return verifiedOperation, receipt, release, authority
}

func retainedAuthorityObject(object cutoverv1.AuthorityAdoptionReceiptObject) cutoverv1.AuthorityAdoptionObject {
	return cutoverv1.AuthorityAdoptionObject{
		Kind: object.Kind, SchemaName: object.SchemaName, ObjectName: object.ObjectName,
		IdentityArguments: object.IdentityArguments, Action: "retain",
		IntroducedGeneration: object.IntroducedGeneration,
		PreAdoptionOwner:     object.Owner, AdoptedOwner: object.Owner,
		PreDefinitionSHA256:     object.DefinitionSHA256,
		AdoptedDefinitionSHA256: object.DefinitionSHA256,
		PreAdoptionACLSHA256:    object.ACLSHA256, AdoptedACLSHA256: object.ACLSHA256,
	}
}

func authorityChainReceipt(
	operation cutoverv1.AuthorityAdoptionOperation, operationDigest, releaseDigest string,
	previous cutoverv1.AuthorityAdoptionReceipt,
	objects []cutoverv1.AuthorityAdoptionReceiptObject,
) cutoverv1.AuthorityAdoptionReceipt {
	receipt := cutoverv1.AuthorityAdoptionReceipt{
		OperationID: operation.OperationID, ReleaseID: operation.ReleaseID,
		EnvironmentID: operation.EnvironmentID, Mode: operation.Mode,
		PreviousCatalogGeneration:      operation.PreviousCatalogGeneration,
		CatalogGeneration:              operation.CatalogGeneration,
		PreCatalogContractSHA256:       operation.PreCatalogContractSHA256,
		PreviousAuthorityReceiptSHA256: cloneStringPointer(operation.PreviousAuthorityReceiptSHA256),
		UpgradeArtifactSHA256:          cloneStringPointer(operation.UpgradeArtifactSHA256),
		Target:                         operation.Target, OperationSHA256: operationDigest,
		MigrationManifestSHA256:     operation.MigrationManifestSHA256,
		ReleaseManifestSHA256:       releaseDigest,
		BootstrapAuthoritySHA256:    operation.BootstrapAuthoritySHA256,
		AdoptedAt:                   fmt.Sprintf("2026-07-14T10:0%d:00Z", operation.CatalogGeneration),
		ExecutorSessionUser:         operation.ExecutorSessionUser,
		PreAdoptionOwnedObjectCount: len(objects) - 1,
		AuthorityOwnerRole:          operation.AuthorityOwnerRole, Objects: objects,
	}
	receipt.PostCatalogContractSHA256 = cutoverv1.AuthorityReceiptCatalogContractSHA256(objects)
	receipt.CatalogInstanceSHA256 = cutoverv1.AuthorityReceiptCatalogInstanceSHA256(objects)
	receipt.QuiescenceEvidence = authorityQuiescenceEvidenceFixture(operation, objects)
	receipt.PreviousCatalogCheckpoint = &cutoverv1.AuthorityCatalogCheckpoint{
		ReceiptSHA256: *operation.PreviousAuthorityReceiptSHA256,
		ReceiptID:     previous.ReceiptID, EnvironmentID: previous.EnvironmentID,
		Target: previous.Target, CatalogGeneration: previous.CatalogGeneration,
		AdoptedAt:             previous.AdoptedAt,
		CatalogContractSHA256: previous.PostCatalogContractSHA256,
		CatalogInstanceSHA256: previous.CatalogInstanceSHA256,
		Objects:               append([]cutoverv1.AuthorityAdoptionReceiptObject(nil), previous.Objects...),
	}
	return receipt
}

func authorityChainTarget() cutoverv1.DatabaseIdentity {
	return cutoverv1.DatabaseIdentity{
		DatabaseName: "harpy", DatabaseOrigin: "fresh",
		SystemIdentifier: "123", DatabaseOID: 16384,
	}
}

func authorityChainDependencies() []cutoverv1.AuthorityCatalogDependency {
	digest := strings.Repeat("b", 64)
	names := []string{
		"platform",
		"cutover_operations", "cutover_operation_relations",
		"cutover_operation_checkpoint_keys", "cutover_operation_columns",
		"cutover_operation_procedures", "cutover_initial_copy_receipts", "cutover_batch_receipts",
		"cutover_copy_checkpoints", "applied_change_receipts", "cutover_change_journal",
		"cutover_credential_receipts",
		"tenant_migration_requirement_catalogs", "tenant_migration_requirements",
		"tenant_migration_requirement_targets",
		"tenant_migration_checkpoints", "tenants", "release_generations", "schema_contracts",
	}
	dependencies := make([]cutoverv1.AuthorityCatalogDependency, len(names))
	for index, name := range names {
		kind, arguments := "table", ""
		if index == 0 {
			kind = "schema"
		}
		dependencies[index] = cutoverv1.AuthorityCatalogDependency{
			Kind: kind, SchemaName: "platform", ObjectName: name,
			IdentityArguments: arguments, Owner: "harpy_owner",
			DefinitionSHA256: digest, ACLSHA256: digest,
		}
	}
	return dependencies
}
