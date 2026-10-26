package environmentrecoveryv1

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"regexp"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/google/uuid"
)

const timeFormat = time.RFC3339Nano

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func validInput(input Input) bool {
	return validUUID(input.EnvironmentID) && validUUID(input.OperationID) &&
		validUUID(input.ArtifactID) && input.Sequence > 0 &&
		canonicalTime(input.IssuedAt) && canonicalTime(input.ExpiresAt) &&
		input.ExpiresAt.After(input.IssuedAt)
}

func validWire(value wire, bundle recoverysigningv1.Bundle, now time.Time) bool {
	purpose, known := signingPurpose(value.Purpose)
	issuedAt, issuedErr := time.Parse(timeFormat, value.IssuedAt)
	expiresAt, expiryErr := time.Parse(timeFormat, value.ExpiresAt)
	payload, payloadErr := canonicaljsonv1.CanonicalizeObject(value.Payload, MaximumPayloadBytes)
	return known && value.Schema == Schema && value.Algorithm == Algorithm &&
		value.EnvironmentID == bundle.EnvironmentID() && validUUID(value.EnvironmentID) &&
		validUUID(value.OperationID) && validUUID(value.ArtifactID) && value.Sequence > 0 &&
		value.KeyGeneration > 0 && value.KeyGeneration <= bundle.Generation() &&
		value.SignerKID != "" && len(value.SignerKID) <= 128 &&
		digestPattern.MatchString(value.PayloadSHA256) && payloadErr == nil &&
		bytes.Equal(payload, value.Payload) && issuedErr == nil && expiryErr == nil &&
		canonicalTime(issuedAt) && canonicalTime(expiresAt) && expiresAt.After(issuedAt) &&
		!now.Before(issuedAt) && now.Before(expiresAt) && purpose != ""
}

func signingPurpose(value Purpose) (recoverysigningv1.Purpose, bool) {
	switch value {
	case PurposeRecoveryManifest:
		return recoverysigningv1.PurposeRecoveryManifest, true
	case PurposeRestorePlan:
		return recoverysigningv1.PurposeRecoveryControl, true
	case PurposeRestoreProgress:
		return recoverysigningv1.PurposeRestoreReceipt, true
	case PurposeRestoreReceipt:
		return recoverysigningv1.PurposeRestoreReceipt, true
	default:
		return "", false
	}
}

func signatureDomain(purpose Purpose, canonical []byte) []byte {
	prefix := []byte("harpy.environment-recovery." + string(purpose) + ".v1")
	return append(append(prefix, 0), canonical...)
}

func validUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed.String() == value
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	return err
}

func PayloadSHA256(payload []byte) string {
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}
