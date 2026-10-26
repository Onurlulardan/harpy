package runtimev1

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const MaximumAttestationDocumentBytes = 64 * 1024

var (
	contractUUIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	digestPattern       = regexp.MustCompile(`^[0-9a-f]{64}$`)
	databaseNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
	keyIDPattern        = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)
	extensionVersion    = regexp.MustCompile(`^[0-9]+\.[0-9]+(?:\.[0-9]+)?$`)
)

type AttestationSignature struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"key_id"`
	Value     string `json:"value"`
}

func ValidateAttestationSignature(value AttestationSignature) error {
	if value.Algorithm != "Ed25519" || !keyIDPattern.MatchString(value.KeyID) {
		return fmt.Errorf("attestation signature identity is invalid")
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value.Value)
	if err != nil || len(decoded) != 64 ||
		base64.RawURLEncoding.EncodeToString(decoded) != value.Value {
		return fmt.Errorf("attestation signature value is invalid")
	}
	return nil
}

func validateUUID(value, field string) error {
	if !contractUUIDPattern.MatchString(value) {
		return fmt.Errorf("%s is invalid", field)
	}
	return nil
}

func validateDigest(value, field string) error {
	if !digestPattern.MatchString(value) {
		return fmt.Errorf("%s is invalid", field)
	}
	return nil
}

func validateTimestamp(value, field string) (time.Time, error) {
	if len(value) == 0 || value[len(value)-1] != 'Z' {
		return time.Time{}, fmt.Errorf("%s is invalid", field)
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s is invalid", field)
	}
	return parsed, nil
}

func validateSystemIdentifier(value string) error {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 || strconv.FormatUint(parsed, 10) != value {
		return fmt.Errorf("database system identifier is invalid")
	}
	return nil
}

func validRoleName(value string) bool { return databaseNamePattern.MatchString(value) }
