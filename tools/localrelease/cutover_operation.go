package main

import (
	"fmt"
	"strings"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func buildCutoverOperation(
	now time.Time, ids releaseIDs, live liveDatabaseIdentity,
	catalog releaseCatalog, authority signedArtifactPair, signer signingAuthority,
) (signedArtifactPair, cutoverv1.AuthorityAdoptionOperation, error) {
	digest := strings.Repeat("a", 64)
	objects := make([]cutoverv1.AuthorityAdoptionObject, len(cutoverObjectOrder))
	for index, identity := range cutoverObjectOrder {
		objects[index] = cutoverv1.AuthorityAdoptionObject{
			Kind: identity[0], SchemaName: identity[1], ObjectName: identity[2],
			IdentityArguments: identity[3], Action: "adopt", IntroducedGeneration: 1,
			PreAdoptionOwner:    "harpy_owner",
			AdoptedOwner:        "harpy_cutover_authority_owner",
			PreDefinitionSHA256: digest, AdoptedDefinitionSHA256: digest,
			PreAdoptionACLSHA256: digest, AdoptedACLSHA256: digest,
		}
	}
	dependencies := make([]cutoverv1.AuthorityCatalogDependency, len(cutoverDependencyOrder))
	for index, identity := range cutoverDependencyOrder {
		dependencies[index] = cutoverv1.AuthorityCatalogDependency{
			Kind: identity[0], SchemaName: identity[1], ObjectName: identity[2],
			IdentityArguments: identity[3], Owner: "harpy_owner",
			DefinitionSHA256: digest, ACLSHA256: digest,
		}
	}
	issuedAt := now.Add(-time.Minute)
	operation := cutoverv1.AuthorityAdoptionOperation{
		ContractVersion: "cutover-authority-adoption-operation.v1",
		OperationID:     ids.cutoverOperation, ReleaseID: ids.release,
		EnvironmentID: ids.environment, Mode: "initial_adoption", CatalogGeneration: 1,
		PreCatalogContractSHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		IssuedAt:                 issuedAt.Format(time.RFC3339),
		ExpiresAt:                now.Add(14 * time.Minute).Format(time.RFC3339),
		Target: cutoverv1.DatabaseIdentity{
			DatabaseName: databaseName, DatabaseOrigin: "fresh",
			SystemIdentifier: live.systemID, DatabaseOID: live.databaseOID,
		},
		MigrationManifestSHA256:  catalog.migration.digest,
		BootstrapAuthoritySHA256: authority.document.digest,
		Executor:                 "external_root_provisioner", ExecutorSessionUser: "postgres",
		OwnershipMode: "superuser_direct_per_object", MembershipPolicy: "always_zero",
		AuthorityOwnerRole: authorityOwnerRole(len(objects)), Objects: objects,
		Dependencies: dependencies,
	}
	operation.PostCatalogContractSHA256 = cutoverv1.AuthorityOperationCatalogContractSHA256(objects)
	operation.DependencyContractSHA256 = cutoverv1.AuthorityDependencyContractSHA256(dependencies)
	operation.QuiescencePolicy = authorityQuiescence(objects)
	if err := operation.Validate(); err != nil {
		return signedArtifactPair{}, operation, fmt.Errorf("validate cutover operation: %w", err)
	}
	document, err := newArtifact("cutover-authority-operation.json", operation)
	if err != nil {
		return signedArtifactPair{}, operation, err
	}
	signatureValue, err := cutoverSignature(signer, document.document, issuedAt)
	if err != nil {
		return signedArtifactPair{}, operation, err
	}
	signature := cutoverv1.AuthorityAdoptionSignature{
		ContractVersion: "cutover-authority-adoption-operation-signature.v1",
		ManifestSHA256:  document.digest, Signature: signatureValue,
	}
	if err := signature.Validate(signature.ContractVersion); err != nil {
		return signedArtifactPair{}, operation, err
	}
	signed, err := newArtifact("cutover-authority-operation.sig", signature)
	return signedArtifactPair{document: document, signature: signed}, operation, err
}

func authorityOwnerRole(objectCount int) cutoverv1.AuthorityOwnerRoleState {
	return cutoverv1.AuthorityOwnerRoleState{
		ConnectionLimit: -1, PasswordState: "null", ValidUntil: "infinity",
		OwnedObjectCount: objectCount,
	}
}

func authorityQuiescence(
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
			Principal: "harpy_migrator", PreState: atRest, TemporaryState: temporary,
			PostState: atRest, PasswordVerifierMustRemainUnchanged: true,
		}},
		TerminateSessions: true, RequireZeroRemainingSessions: true,
		RestoreBeforeCommit:         true,
		ExpectedPreNamespaceSHA256:  cutoverv1.AuthorityOperationPreNamespaceSHA256(objects),
		ExpectedPostNamespaceSHA256: cutoverv1.AuthorityOperationPostNamespaceSHA256(objects),
	}
}
