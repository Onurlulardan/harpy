package main

import (
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func buildCutoverReceipt(
	now time.Time, ids releaseIDs, operation cutoverv1.AuthorityAdoptionOperation,
	operationArtifacts, authority, release signedArtifactPair, signer signingAuthority,
) (signedArtifactPair, error) {
	objects := make([]cutoverv1.AuthorityAdoptionReceiptObject, len(operation.Objects))
	for index, object := range operation.Objects {
		objects[index] = cutoverv1.AuthorityAdoptionReceiptObject{
			Kind: object.Kind, SchemaName: object.SchemaName, ObjectName: object.ObjectName,
			IdentityArguments:    object.IdentityArguments,
			IntroducedGeneration: object.IntroducedGeneration, ObjectOID: uint32(70000 + index),
			Owner: object.AdoptedOwner, DefinitionSHA256: object.AdoptedDefinitionSHA256,
			ACLSHA256: object.AdoptedACLSHA256,
		}
	}
	dependencies := make([]cutoverv1.AuthorityCatalogDependencyReceipt, len(operation.Dependencies))
	for index, dependency := range operation.Dependencies {
		dependencies[index] = cutoverv1.AuthorityCatalogDependencyReceipt{
			AuthorityCatalogDependency: dependency, ObjectOID: uint32(71000 + index),
		}
	}
	receipt := cutoverv1.AuthorityAdoptionReceipt{
		ContractVersion: "cutover-authority-adoption-receipt.v1",
		ReceiptID:       ids.cutoverReceipt, OperationID: operation.OperationID,
		ReleaseID: operation.ReleaseID, EnvironmentID: operation.EnvironmentID,
		Mode: operation.Mode, PreviousCatalogGeneration: operation.PreviousCatalogGeneration,
		CatalogGeneration:         operation.CatalogGeneration,
		PreCatalogContractSHA256:  operation.PreCatalogContractSHA256,
		PostCatalogContractSHA256: operation.PostCatalogContractSHA256,
		Target:                    operation.Target, OperationSHA256: operationArtifacts.document.digest,
		MigrationManifestSHA256:  operation.MigrationManifestSHA256,
		ReleaseManifestSHA256:    release.document.digest,
		BootstrapAuthoritySHA256: authority.document.digest,
		AdoptedAt:                now.Add(-30 * time.Second).Format(time.RFC3339), Result: "succeeded",
		ExecutorSessionUser: operation.ExecutorSessionUser, ExecutorSuperuser: true,
		AuthorityOwnerRole: operation.AuthorityOwnerRole, Objects: objects,
		DependencyContractSHA256: operation.DependencyContractSHA256,
		Dependencies:             dependencies,
	}
	receipt.PostCatalogContractSHA256 = cutoverv1.AuthorityReceiptCatalogContractSHA256(objects)
	receipt.CatalogInstanceSHA256 = cutoverv1.AuthorityReceiptCatalogInstanceSHA256(objects)
	receipt.DependencyContractSHA256 = cutoverv1.AuthorityReceiptDependencyContractSHA256(dependencies)
	receipt.DependencyInstanceSHA256 = cutoverv1.AuthorityReceiptDependencyInstanceSHA256(dependencies)
	receipt.QuiescenceEvidence = authorityQuiescenceEvidence(operation, objects)
	if err := receipt.Validate(); err != nil {
		return signedArtifactPair{}, fmt.Errorf("validate cutover receipt: %w", err)
	}
	document, err := newArtifact("cutover-authority-receipt.json", receipt)
	if err != nil {
		return signedArtifactPair{}, err
	}
	signatureValue, err := cutoverSignature(signer, document.document, now.Add(-30*time.Second))
	if err != nil {
		return signedArtifactPair{}, err
	}
	signature := cutoverv1.AuthorityAdoptionSignature{
		ContractVersion: "cutover-authority-adoption-receipt-signature.v1",
		ManifestSHA256:  document.digest, Signature: signatureValue,
	}
	if err := signature.Validate(signature.ContractVersion); err != nil {
		return signedArtifactPair{}, err
	}
	signed, err := newArtifact("cutover-authority-receipt.sig", signature)
	return signedArtifactPair{document: document, signature: signed}, err
}

func authorityQuiescenceEvidence(
	operation cutoverv1.AuthorityAdoptionOperation,
	objects []cutoverv1.AuthorityAdoptionReceiptObject,
) cutoverv1.AuthorityQuiescenceEvidence {
	policy := operation.QuiescencePolicy.Roles[0]
	return cutoverv1.AuthorityQuiescenceEvidence{
		Roles: []cutoverv1.AuthorityQuiescedRoleEvidence{{
			Principal: policy.Principal, ObservedPreState: policy.PreState,
			ObservedTemporaryState: policy.TemporaryState, ObservedPostState: policy.PostState,
			PasswordVerifierUnchanged: true,
		}},
		ObservedPreNamespaceSHA256:  operation.QuiescencePolicy.ExpectedPreNamespaceSHA256,
		ObservedPostNamespaceSHA256: cutoverv1.AuthorityReceiptNamespaceSHA256(objects),
		RestoredBeforeCommit:        true,
	}
}
