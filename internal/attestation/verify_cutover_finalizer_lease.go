package attestation

import (
	"crypto/subtle"
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func VerifyCutoverFinalizerLease(
	manifestDocument []byte,
	signatureDocument []byte,
	adoption VerifiedCutoverCredentialLease,
	cleanup VerifiedCutoverCredentialCleanup,
	authorityReceipt VerifiedAuthorityAdoptionReceipt,
	root TrustRoot,
	now time.Time,
) (VerifiedCutoverFinalizerLease, error) {
	if !adoption.verified || !cleanup.verified || !authorityReceipt.verified {
		return VerifiedCutoverFinalizerLease{}, fmt.Errorf("cutover finalizer prerequisites were not verified")
	}
	var lease cutoverv1.FinalizerLease
	if err := DecodeStrict(manifestDocument, &lease); err != nil {
		return VerifiedCutoverFinalizerLease{}, fmt.Errorf("decode cutover finalizer lease: %w", err)
	}
	if err := cutoverv1.ValidateFinalizerLease(lease); err != nil {
		return VerifiedCutoverFinalizerLease{}, err
	}
	issuedAt, _ := parseCutoverTime(lease.IssuedAt)
	notBefore, _ := parseCutoverTime(lease.NotBefore)
	expiresAt, _ := parseCutoverTime(lease.ExpiresAt)
	if now.Before(notBefore) || !now.Before(expiresAt) {
		return VerifiedCutoverFinalizerLease{}, fmt.Errorf("cutover finalizer lease is not active")
	}
	var signature cutoverv1.FinalizerLeaseSignature
	if err := DecodeStrict(signatureDocument, &signature); err != nil {
		return VerifiedCutoverFinalizerLease{}, fmt.Errorf("decode cutover finalizer signature: %w", err)
	}
	if err := cutoverv1.ValidateFinalizerLeaseSignature(signature); err != nil {
		return VerifiedCutoverFinalizerLease{}, err
	}
	digest, err := verifyFinalizerSignature(manifestDocument, signature, root, issuedAt)
	if err != nil {
		return VerifiedCutoverFinalizerLease{}, err
	}
	if lease.OperationID != adoption.OperationID() || lease.RunID != adoption.RunID() ||
		lease.ReleaseID != adoption.ReleaseID() || lease.EnvironmentID != adoption.EnvironmentID() ||
		lease.Target != adoption.Target() || lease.AdoptionLeaseSHA256 != adoption.ManifestSHA256() ||
		lease.CleanupReceiptSHA256 != cleanup.ReceiptSHA256() ||
		lease.AuthorityAdoptionReceiptSHA256 != authorityReceipt.ReceiptSHA256() ||
		!finalizerAuthorityReceiptMatches(lease, authorityReceipt.receipt) {
		return VerifiedCutoverFinalizerLease{}, fmt.Errorf("cutover finalizer lease binding differs")
	}
	canonical, err := Canonicalize(manifestDocument)
	if err != nil {
		return VerifiedCutoverFinalizerLease{}, err
	}
	return VerifiedCutoverFinalizerLease{
		lease: lease, manifestSHA256: digest, canonical: canonical,
		expiresAt: expiresAt, verified: true,
	}, nil
}

func finalizerAuthorityReceiptMatches(
	lease cutoverv1.FinalizerLease,
	receipt cutoverv1.AuthorityAdoptionReceipt,
) bool {
	if receipt.ReleaseID != lease.ReleaseID || receipt.EnvironmentID != lease.EnvironmentID ||
		receipt.Target != lease.Target || receipt.Result != "succeeded" {
		return false
	}
	for _, object := range receipt.Objects {
		if object.SchemaName == "cutover_authority" &&
			object.ObjectName == "finalize_operation_v1" &&
			object.IdentityArguments ==
				"uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea" {
			return object.ObjectOID == lease.FinalizationProcedureOID
		}
	}
	return false
}

func verifyFinalizerSignature(
	document []byte,
	signed cutoverv1.FinalizerLeaseSignature,
	root TrustRoot,
	issuedAt time.Time,
) (string, error) {
	digest, err := DigestHex(document)
	if err != nil {
		return "", err
	}
	if subtle.ConstantTimeCompare([]byte(digest), []byte(signed.ManifestSHA256)) != 1 {
		return "", fmt.Errorf("cutover finalizer lease digest differs")
	}
	if err := root.Verify(document, cutoverSignature(signed.Signature), issuedAt); err != nil {
		return "", fmt.Errorf("verify cutover finalizer lease signature: %w", err)
	}
	return digest, nil
}
