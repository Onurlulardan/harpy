package iamrestorev1

import (
	"encoding/base64"
	"regexp"
	"slices"

	"github.com/google/uuid"
)

var (
	digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	handlePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/@-]{0,254}$`)
)

func validSecurity(value SecurityEvidence) bool {
	return value.Schema == SecurityEvidenceSchema && validUUIDs(
		value.EnvironmentID, value.RestoreOperationID, value.RecoverySetID,
		value.TenantID, value.OwnerUserID, value.OwnerBindingID, value.AnchorID,
		value.NewDatabaseSystemID,
	) && value.OwnerBindingGeneration > 0 && validDigests(
		value.RestoreReceiptSHA256, value.RecoveryTrustSHA256, value.AckTrustSHA256,
		value.CustodyManifestSHA256, value.ProtectedHeadSHA256,
		value.ProtectedAckSHA256, value.ProtectedCheckpointSHA256,
	) && len(value.ProtectedAck) > 1 && len(value.ProtectedCheckpoint) > 1
}

func validCustody(value CustodyManifest) bool {
	if value.Schema != CustodyManifestSchema || !validUUIDs(value.EnvironmentID) ||
		!handlePattern.MatchString(value.Provider) ||
		!handlePattern.MatchString(value.KeyringHandle) || len(value.Rotations) == 0 {
		return false
	}
	previous := ""
	for _, rotation := range value.Rotations {
		if !validRotation(rotation) || previous >= rotation.OperationID {
			return false
		}
		previous = rotation.OperationID
	}
	return true
}

func validRotation(value RotationResult) bool {
	if !validUUIDs(value.OperationID) || value.Status != "rotated" ||
		value.Generation < 1 || !validDigests(value.ResultSHA256) ||
		!validSecurityEpoch(value.SecurityEpoch) || rotationDigest(value) != value.ResultSHA256 ||
		len(value.PublicKIDs) == 0 || !slices.IsSorted(value.PublicKIDs) ||
		len(value.RevokedGenerations) == 0 || !slices.IsSorted(value.RevokedGenerations) {
		return false
	}
	for index, kid := range value.PublicKIDs {
		if !validKID(kid) || index > 0 && value.PublicKIDs[index-1] == kid {
			return false
		}
	}
	for index, generation := range value.RevokedGenerations {
		if generation < 1 || generation >= value.Generation ||
			index > 0 && value.RevokedGenerations[index-1] == generation {
			return false
		}
	}
	return true
}

func validSecurityEpoch(value string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) != 32 ||
		base64.RawURLEncoding.EncodeToString(decoded) != value {
		return false
	}
	var nonZero byte
	for _, item := range decoded {
		nonZero |= item
	}
	return nonZero != 0
}

func validKID(value string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(decoded) >= 16 && len(decoded) <= 64
}

func validUUIDs(values ...string) bool {
	for _, value := range values {
		parsed, err := uuid.Parse(value)
		if err != nil || parsed == uuid.Nil || parsed.String() != value {
			return false
		}
	}
	return true
}

func validDigests(values ...string) bool {
	return !slices.ContainsFunc(values, func(value string) bool {
		return !digestPattern.MatchString(value)
	})
}
