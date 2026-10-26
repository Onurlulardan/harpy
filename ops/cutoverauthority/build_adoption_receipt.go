package main

import (
	"context"
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/jackc/pgx/v5"
)

type signedAdoptionReceipt struct {
	receipt           cutoverv1.AuthorityAdoptionReceipt
	document          []byte
	signatureDocument []byte
	digest            string
}

func buildSignedAdoptionReceipt(
	ctx context.Context, transaction pgx.Tx, bundle verifiedAuthorityBundle,
	signer attestation.Signer, receiptID string, adoptedAt time.Time,
	postStates, dependencyStates []catalogState, quiescence quiescenceRun,
) (signedAdoptionReceipt, error) {
	operation := bundle.operation.Value()
	previous, err := loadPreviousCheckpoint(ctx, transaction, operation)
	if err != nil {
		return signedAdoptionReceipt{}, err
	}
	objects := make([]cutoverv1.AuthorityAdoptionReceiptObject, len(operation.Objects))
	for index, object := range operation.Objects {
		state := postStates[index]
		objects[index] = cutoverv1.AuthorityAdoptionReceiptObject{
			Kind: object.Kind, SchemaName: object.SchemaName, ObjectName: object.ObjectName,
			IdentityArguments:    object.IdentityArguments,
			IntroducedGeneration: object.IntroducedGeneration, ObjectOID: state.oid,
			Owner: state.owner, DefinitionSHA256: state.definition, ACLSHA256: state.acl,
		}
	}
	dependencies := make([]cutoverv1.AuthorityCatalogDependencyReceipt,
		len(operation.Dependencies))
	for index, dependency := range operation.Dependencies {
		dependencies[index] = cutoverv1.AuthorityCatalogDependencyReceipt{
			AuthorityCatalogDependency: dependency, ObjectOID: dependencyStates[index].oid,
		}
	}
	postRole, err := readQuiescedRole(ctx, transaction, quiescence.policy.Principal)
	if err != nil || postRole.state != quiescence.policy.PostState ||
		postRole.verifier != quiescence.pre.verifier {
		return signedAdoptionReceipt{}, fmt.Errorf("restored migrator role differs")
	}
	receipt := cutoverv1.AuthorityAdoptionReceipt{
		ContractVersion: "cutover-authority-adoption-receipt.v1",
		ReceiptID:       receiptID, OperationID: operation.OperationID,
		ReleaseID: operation.ReleaseID, EnvironmentID: operation.EnvironmentID,
		Mode: operation.Mode, PreviousCatalogGeneration: operation.PreviousCatalogGeneration,
		CatalogGeneration:              operation.CatalogGeneration,
		PreCatalogContractSHA256:       operation.PreCatalogContractSHA256,
		PostCatalogContractSHA256:      operation.PostCatalogContractSHA256,
		PreviousAuthorityReceiptSHA256: operation.PreviousAuthorityReceiptSHA256,
		UpgradeArtifactSHA256:          operation.UpgradeArtifactSHA256,
		PreviousCatalogCheckpoint:      previous, Target: operation.Target,
		OperationSHA256:          bundle.operation.ManifestSHA256(),
		MigrationManifestSHA256:  operation.MigrationManifestSHA256,
		ReleaseManifestSHA256:    bundle.release.ManifestSHA256(),
		BootstrapAuthoritySHA256: bundle.authority.ManifestSHA256(),
		AdoptedAt:                adoptedAt.UTC().Format(time.RFC3339Nano), Result: "succeeded",
		ExecutorSessionUser: operation.ExecutorSessionUser, ExecutorSuperuser: true,
		PreAdoptionMembershipCount:    0,
		PreAdoptionOwnedObjectCount:   previousObjectCount(operation),
		AuthorityOwnerMembershipCount: 0, AuthorityOwnerRole: operation.AuthorityOwnerRole,
		Objects: objects, DependencyContractSHA256: operation.DependencyContractSHA256,
		Dependencies: dependencies,
	}
	receipt.QuiescenceEvidence = cutoverv1.AuthorityQuiescenceEvidence{
		Roles: []cutoverv1.AuthorityQuiescedRoleEvidence{{
			Principal:              quiescence.policy.Principal,
			ObservedPreState:       quiescence.policy.PreState,
			ObservedTemporaryState: quiescence.temporary.state,
			ObservedPostState:      postRole.state,
			TerminatedSessionCount: quiescence.terminated, RemainingSessionCount: 0,
			PasswordVerifierUnchanged: postRole.verifier == quiescence.pre.verifier,
		}},
		ObservedPreNamespaceSHA256:  operation.QuiescencePolicy.ExpectedPreNamespaceSHA256,
		ObservedPostNamespaceSHA256: cutoverv1.AuthorityReceiptNamespaceSHA256(objects),
		RestoredBeforeCommit:        true,
	}
	receipt.CatalogInstanceSHA256 = cutoverv1.AuthorityReceiptCatalogInstanceSHA256(objects)
	receipt.DependencyInstanceSHA256 =
		cutoverv1.AuthorityReceiptDependencyInstanceSHA256(dependencies)
	return signAdoptionReceipt(receipt, bundle, signer, adoptedAt)
}

func previousObjectCount(operation cutoverv1.AuthorityAdoptionOperation) int {
	count := 0
	for _, object := range operation.Objects {
		if object.IntroducedGeneration <= operation.PreviousCatalogGeneration {
			count++
		}
	}
	return count
}

func signAdoptionReceipt(
	receipt cutoverv1.AuthorityAdoptionReceipt, bundle verifiedAuthorityBundle,
	signer attestation.Signer, adoptedAt time.Time,
) (signedAdoptionReceipt, error) {
	if err := receipt.Validate(); err != nil {
		return signedAdoptionReceipt{}, fmt.Errorf("authority receipt validation failed")
	}
	document, err := attestation.MarshalCanonical(receipt)
	if err != nil {
		return signedAdoptionReceipt{}, fmt.Errorf("authority receipt canonicalization failed")
	}
	digest, err := attestation.DigestHex(document)
	if err != nil {
		return signedAdoptionReceipt{}, fmt.Errorf("authority receipt digest failed")
	}
	signed, err := signer.SignAt(document, adoptedAt)
	if err != nil {
		return signedAdoptionReceipt{}, fmt.Errorf("authority receipt signing failed")
	}
	signature, err := attestation.MarshalCanonical(cutoverv1.AuthorityAdoptionSignature{
		ContractVersion: "cutover-authority-adoption-receipt-signature.v1",
		ManifestSHA256:  digest, Signature: cutoverv1.Signature{
			Algorithm: signed.Algorithm, KeyID: signed.KeyID, Value: signed.Value,
		},
	})
	if err != nil {
		return signedAdoptionReceipt{}, fmt.Errorf("authority receipt signature encoding failed")
	}
	if _, err := attestation.VerifyAuthorityAdoptionReceipt(document, signature,
		bundle.operation, bundle.release, bundle.authority, bundle.root, adoptedAt); err != nil {
		return signedAdoptionReceipt{}, fmt.Errorf("authority receipt self-verification failed: %w", err)
	}
	return signedAdoptionReceipt{
		receipt: receipt, document: document, signatureDocument: signature, digest: digest,
	}, nil
}
