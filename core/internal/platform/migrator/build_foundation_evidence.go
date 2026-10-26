package migrator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func buildFoundationEvidence(
	artifacts attestation.VerifiedPostgresArtifactSet,
	bootstrap runtimev1.BootstrapReceipt,
	extension runtimev1.ExtensionReceipt,
	cleanup runtimev1.CredentialCleanupReceipt,
) (foundationEvidence, error) {
	receipts := make([]receiptEvidence, 0, 3)
	bootstrapEvidence, err := makeReceiptEvidence(
		bootstrap.Kind, bootstrap.Payload.RunID, bootstrap.Payload.OperationDigest,
		bootstrap.Payload.IssuedAt, bootstrap.Signature.KeyID, bootstrap,
	)
	if err != nil {
		return foundationEvidence{}, err
	}
	receipts = append(receipts, bootstrapEvidence)
	extensionEvidence, err := makeReceiptEvidence(
		extension.Kind, extension.Payload.RunID, extension.Payload.OperationDigest,
		extension.Payload.IssuedAt, extension.Signature.KeyID, extension,
	)
	if err != nil {
		return foundationEvidence{}, err
	}
	receipts = append(receipts, extensionEvidence)
	cleanupEvidence, err := makeReceiptEvidence(
		cleanup.Kind, cleanup.Payload.RunID, cleanup.Payload.CleanupOperationDigest,
		cleanup.Payload.IssuedAt, cleanup.Signature.KeyID, cleanup,
	)
	if err != nil {
		return foundationEvidence{}, err
	}
	receipts = append(receipts, cleanupEvidence)
	extensions := makeExtensionEvidence(
		artifacts.Artifacts().Extension.Value().Extensions,
		extension.Payload.RunID,
	)
	return foundationEvidence{receipts: receipts, extensions: extensions}, nil
}

func makeReceiptEvidence(
	kind, runID, operationDigest, issuedAt, keyID string,
	document any,
) (receiptEvidence, error) {
	canonical, err := attestation.MarshalCanonical(document)
	if err != nil {
		return receiptEvidence{}, fmt.Errorf("canonicalize %s receipt: %w", kind, err)
	}
	parsedIssuedAt, err := time.Parse(time.RFC3339Nano, issuedAt)
	if err != nil {
		return receiptEvidence{}, fmt.Errorf("parse %s receipt time: %w", kind, err)
	}
	digest := sha256.Sum256(canonical)
	return receiptEvidence{
		kind: kind, runID: runID, operationDigest: operationDigest,
		canonicalDocument: canonical, canonicalSHA256: hex.EncodeToString(digest[:]),
		signatureKeyID: keyID, issuedAt: parsedIssuedAt,
	}, nil
}

func makeExtensionEvidence(
	extensions []postgresv1.PostgresExtension,
	verifiedRunID string,
) []extensionEvidence {
	evidence := make([]extensionEvidence, 0, len(extensions))
	for _, extension := range extensions {
		evidence = append(evidence, extensionEvidence{
			name: extension.Name, version: extension.Version,
			schemaName: extension.Schema, ownerRole: extension.Owner,
			controlFileSHA256: extension.ControlFileSHA256,
			capabilities:      append([]string(nil), extension.Capabilities...),
			verifiedRunID:     verifiedRunID,
		})
	}
	return evidence
}
