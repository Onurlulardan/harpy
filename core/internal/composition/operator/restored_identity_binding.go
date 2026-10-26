package operator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformopsmodule "github.com/Onurlulardan/harpy/core/internal/modules/platformops"
	"github.com/Onurlulardan/harpy/core/internal/platform/recoveryartifact"
)

func bindRestoredIdentity(
	receiptDocument, trustDocument []byte,
	receiptArtifact recoveryartifact.Artifact,
	receipt platformopsmodule.RestoreReceiptPayload,
	security iammodule.RestoredSecurityEvidence,
) (restoredIdentityArtifacts, error) {
	operation, operationErr := iammodule.ParseUUID(receiptArtifact.OperationID())
	environment, environmentErr := iammodule.ParseUUID(receipt.Identity.EnvironmentID)
	recoverySet, recoveryErr := iammodule.ParseUUID(receipt.RecoverySetID)
	tenant, tenantErr := iammodule.ParseUUID(security.TenantID)
	owner, ownerErr := iammodule.ParseUUID(security.OwnerUserID)
	anchor, anchorErr := iammodule.ParseUUID(security.AnchorID)
	newDatabase, databaseErr := iammodule.ParseUUID(receipt.NewDatabaseSystemID)
	receiptDigest := sha256.Sum256(receiptDocument)
	protectedHead, headErr := decodeDigest(security.ProtectedHeadSHA256)
	valid := operationErr == nil && environmentErr == nil && recoveryErr == nil &&
		tenantErr == nil && ownerErr == nil && anchorErr == nil && databaseErr == nil &&
		headErr == nil && receiptArtifact.EnvironmentID() == receipt.Identity.EnvironmentID &&
		security.EnvironmentID == receipt.Identity.EnvironmentID &&
		security.RestoreOperationID == operation.String() &&
		security.RecoverySetID == recoverySet.String() &&
		security.TenantID == receipt.Identity.PrimaryBindingID &&
		security.OwnerBindingID == receipt.Identity.OwnerBindingID &&
		security.OwnerBindingGeneration == receipt.Identity.OwnerBindingGeneration &&
		security.NewDatabaseSystemID == receipt.NewDatabaseSystemID &&
		security.RestoreReceiptSHA256 == hex.EncodeToString(receiptDigest[:]) &&
		security.RecoveryTrustSHA256 == digestHex(trustDocument)
	if !valid {
		return restoredIdentityArtifacts{}, fmt.Errorf("restored identity binding is invalid")
	}
	return restoredIdentityArtifacts{receiptDocument: append([]byte(nil), receiptDocument...),
		receipt: receipt, restoreOperation: operation, environment: environment,
		recoverySet: recoverySet, security: security, tenant: tenant, owner: owner,
		anchor: anchor, newDatabase: newDatabase, protectedHead: protectedHead,
		receiptSHA256: receiptDigest}, nil
}

func decodeDigest(value string) ([32]byte, error) {
	var result [32]byte
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != len(result) {
		return result, fmt.Errorf("digest is invalid")
	}
	copy(result[:], decoded)
	return result, nil
}
