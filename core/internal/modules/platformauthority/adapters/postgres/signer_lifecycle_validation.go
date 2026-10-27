package postgres

import (
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
	"github.com/google/uuid"
)

func validSignerRotation(value ports.SignerRotationRecord) bool {
	return validIdentifier(value.OperationID) && validIdentifier(value.BindingID) &&
		value.ExpectedSignerGeneration > 0 &&
		value.NewSignerGeneration == value.ExpectedSignerGeneration+1 &&
		value.ExpectedSignerKID != "" && value.NewSignerKID != "" &&
		value.ExpectedSignerKID != value.NewSignerKID &&
		validDigest(value.NewSignerPublicKeySHA256) &&
		(value.NewRecoveryTrustBundleSHA256 == nil ||
			validDigest(*value.NewRecoveryTrustBundleSHA256))
}

func validSignerFence(value ports.SignerFenceRecord) bool {
	return validIdentifier(value.BindingID) && value.ExpectedSignerGeneration > 0 &&
		value.ExpectedSignerKID != "" && validDigest(value.EvidenceSHA256)
}

func validIdentifier(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed.String() == value
}

func validDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32 && hex.EncodeToString(decoded) == value
}
