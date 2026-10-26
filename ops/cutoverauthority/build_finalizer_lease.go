package main

import (
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type signedFinalizerLease struct {
	document          []byte
	signatureDocument []byte
	digest            string
}

func buildSignedFinalizerLease(
	leaseID string, prerequisites verifiedFinalizationPrerequisites,
	observation finalizerAuthorityObservation, signer attestation.Signer,
) (signedFinalizerLease, error) {
	issuedAt := observation.finalizerValidUntil.Add(-120 * time.Second).UTC()
	var tenantCatalogSHA256 *string
	var tenantRequirementCount *int
	if prerequisites.terminal.value.TerminalState == "completed" {
		tenantCatalogSHA256 = observation.tenantCatalogSHA256
		tenantRequirementCount = observation.tenantRequirementCount
	}
	lease := cutoverv1.FinalizerLease{
		ContractVersion: "cutover-finalizer-lease.v1",
		LeaseID:         leaseID, ReleaseID: prerequisites.adoption.ReleaseID(),
		EnvironmentID:      prerequisites.adoption.EnvironmentID(),
		OperationID:        prerequisites.adoption.OperationID(),
		RunID:              prerequisites.adoption.RunID(),
		IssuedAt:           issuedAt.Format(time.RFC3339Nano),
		NotBefore:          issuedAt.Format(time.RFC3339Nano),
		ExpiresAt:          observation.finalizerValidUntil.UTC().Format(time.RFC3339Nano),
		MaxDurationSeconds: 120,
		RoleLease: cutoverv1.RoleLease{Principal: "harpy_cutover_finalizer",
			Login: true, ConnectionLimit: 1, PasswordState: "leased_secret"},
		Target:                         prerequisites.adoption.Target(),
		AdoptionLeaseSHA256:            prerequisites.adoption.ManifestSHA256(),
		CleanupReceiptSHA256:           prerequisites.cleanup.ReceiptSHA256(),
		AuthorityAdoptionReceiptSHA256: prerequisites.receipt.ReceiptSHA256(),
		FinalizationStateSHA256:        observation.stateSHA256,
		TenantRequirementCatalogSHA256: tenantCatalogSHA256,
		TenantRequirementCount:         tenantRequirementCount,
		FinalizationProcedureOID:       observation.procedureOID,
		FinalizationProcedureSignature: observation.procedureSignature,
		TerminalState:                  prerequisites.terminal.value.TerminalState,
		TerminalEvidenceSHA256:         prerequisites.terminal.digest,
	}
	if err := cutoverv1.ValidateFinalizerLease(lease); err != nil {
		return signedFinalizerLease{}, fmt.Errorf("finalizer lease validation failed")
	}
	document, err := attestation.MarshalCanonical(lease)
	if err != nil {
		return signedFinalizerLease{}, fmt.Errorf("finalizer lease encoding failed")
	}
	digest, err := attestation.DigestHex(document)
	if err != nil {
		return signedFinalizerLease{}, fmt.Errorf("finalizer lease digest failed")
	}
	signed, err := signer.SignAt(document, issuedAt)
	if err != nil {
		return signedFinalizerLease{}, fmt.Errorf("finalizer lease signing failed")
	}
	signature, err := attestation.MarshalCanonical(cutoverv1.FinalizerLeaseSignature{
		ContractVersion: "cutover-finalizer-lease-signature.v1",
		ManifestSHA256:  digest, Signature: cutoverv1.Signature{
			Algorithm: signed.Algorithm, KeyID: signed.KeyID, Value: signed.Value,
		},
	})
	if err != nil {
		return signedFinalizerLease{}, fmt.Errorf("finalizer signature encoding failed")
	}
	return signedFinalizerLease{
		document: document, signatureDocument: signature, digest: digest,
	}, nil
}
