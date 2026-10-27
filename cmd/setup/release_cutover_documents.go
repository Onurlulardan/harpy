package main

import (
	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func buildCutoverOperation(
	keys releaseKeys, migrationDigest, authorityDigest string,
) (releaseDocument, []byte, error) {
	objects := cutoverObjects()
	dependencies := cutoverDependencies()
	atRest := cutoverMigratorAtRest()
	temporary := atRest
	temporary.Login = false
	temporary.ConnectionLimit = 0
	operation := cutoverv1.AuthorityAdoptionOperation{
		ContractVersion: "cutover-authority-adoption-operation.v1",
		OperationID:     localOperationID, ReleaseID: localReleaseID,
		EnvironmentID: localEnvironmentID, Mode: "initial_adoption", CatalogGeneration: 1,
		PreCatalogContractSHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		IssuedAt:                 keys.clock.stamp(keys.clock.issued),
		ExpiresAt:                keys.clock.stamp(keys.clock.expires),
		Target: cutoverv1.DatabaseIdentity{
			DatabaseName: localDatabase, DatabaseOrigin: "fresh",
			SystemIdentifier: localSystemID, DatabaseOID: localDatabaseOID,
		},
		MigrationManifestSHA256: migrationDigest, BootstrapAuthoritySHA256: authorityDigest,
		Executor: "external_root_provisioner", ExecutorSessionUser: "postgres",
		OwnershipMode: "superuser_direct_per_object", MembershipPolicy: "always_zero",
		AuthorityOwnerRole:       cutoverOwnerRole(len(objects)),
		Objects:                  objects,
		Dependencies:             dependencies,
		DependencyContractSHA256: cutoverv1.AuthorityDependencyContractSHA256(dependencies),
		QuiescencePolicy: cutoverv1.AuthorityQuiescencePolicy{
			Roles: []cutoverv1.AuthorityQuiescedRolePolicy{{
				Principal: "harpy_migrator", PreState: atRest, TemporaryState: temporary,
				PostState: atRest, PasswordVerifierMustRemainUnchanged: true,
			}},
			TerminateSessions: true, RequireZeroRemainingSessions: true, RestoreBeforeCommit: true,
			ExpectedPreNamespaceSHA256:  cutoverv1.AuthorityOperationPreNamespaceSHA256(objects),
			ExpectedPostNamespaceSHA256: cutoverv1.AuthorityOperationPostNamespaceSHA256(objects),
		},
	}
	operation.PostCatalogContractSHA256 = cutoverv1.AuthorityOperationCatalogContractSHA256(objects)
	if err := operation.Validate(); err != nil {
		return releaseDocument{}, nil, err
	}
	document, err := canonicalDocument(operation)
	if err != nil {
		return releaseDocument{}, nil, err
	}
	signature, err := keys.signDetached(
		"cutover-authority-adoption-operation-signature.v1", document.body,
	)
	if err != nil {
		return releaseDocument{}, nil, err
	}
	return document, signature, nil
}

func verifyCutoverOperation(
	keys releaseKeys, document, signature []byte,
	release attestation.VerifiedRelease, authority attestation.VerifiedBootstrapAuthority,
) (attestation.VerifiedAuthorityAdoptionOperation, error) {
	return attestation.VerifyAuthorityAdoptionOperation(
		document, signature, release, authority, keys.root, keys.clock.now,
	)
}

func buildCutoverReceipt(
	keys releaseKeys, operation cutoverv1.AuthorityAdoptionOperation,
	verified attestation.VerifiedAuthorityAdoptionOperation,
	release attestation.VerifiedRelease, authority attestation.VerifiedBootstrapAuthority,
) ([]byte, []byte, error) {
	objects := cutoverReceiptObjects(operation.Objects)
	dependencies := cutoverReceiptDependencies(operation.Dependencies)
	policy := operation.QuiescencePolicy.Roles[0]
	receipt := cutoverv1.AuthorityAdoptionReceipt{
		ContractVersion: "cutover-authority-adoption-receipt.v1", ReceiptID: localReceiptID,
		OperationID: operation.OperationID, ReleaseID: operation.ReleaseID,
		EnvironmentID: operation.EnvironmentID, Target: operation.Target, Mode: operation.Mode,
		PreviousCatalogGeneration: operation.PreviousCatalogGeneration,
		CatalogGeneration:         operation.CatalogGeneration,
		PreCatalogContractSHA256:  operation.PreCatalogContractSHA256,
		PostCatalogContractSHA256: cutoverv1.AuthorityReceiptCatalogContractSHA256(objects),
		OperationSHA256:           verified.ManifestSHA256(),
		MigrationManifestSHA256:   operation.MigrationManifestSHA256,
		ReleaseManifestSHA256:     release.ManifestSHA256(),
		BootstrapAuthoritySHA256:  authority.ManifestSHA256(),
		AdoptedAt:                 keys.clock.stamp(keys.clock.adopted),
		Result:                    "succeeded", ExecutorSessionUser: operation.ExecutorSessionUser,
		ExecutorSuperuser: true, AuthorityOwnerRole: operation.AuthorityOwnerRole,
		CatalogInstanceSHA256:    cutoverv1.AuthorityReceiptCatalogInstanceSHA256(objects),
		Objects:                  objects,
		Dependencies:             dependencies,
		DependencyContractSHA256: cutoverv1.AuthorityReceiptDependencyContractSHA256(dependencies),
		DependencyInstanceSHA256: cutoverv1.AuthorityReceiptDependencyInstanceSHA256(dependencies),
		QuiescenceEvidence: cutoverv1.AuthorityQuiescenceEvidence{
			Roles: []cutoverv1.AuthorityQuiescedRoleEvidence{{
				Principal: policy.Principal, ObservedPreState: policy.PreState,
				ObservedTemporaryState: policy.TemporaryState, ObservedPostState: policy.PostState,
				PasswordVerifierUnchanged: true,
			}},
			ObservedPreNamespaceSHA256:  operation.QuiescencePolicy.ExpectedPreNamespaceSHA256,
			ObservedPostNamespaceSHA256: cutoverv1.AuthorityReceiptNamespaceSHA256(objects),
			RestoredBeforeCommit:        true,
		},
	}
	if err := receipt.Validate(); err != nil {
		return nil, nil, err
	}
	document, err := canonicalDocument(receipt)
	if err != nil {
		return nil, nil, err
	}
	signature, err := keys.signDetached(
		"cutover-authority-adoption-receipt-signature.v1", document.body,
	)
	if err != nil {
		return nil, nil, err
	}
	if _, err = attestation.VerifyAuthorityAdoptionReceipt(
		document.body, signature, verified, release, authority, keys.root, keys.clock.now,
	); err != nil {
		return nil, nil, err
	}
	return document.body, signature, nil
}
