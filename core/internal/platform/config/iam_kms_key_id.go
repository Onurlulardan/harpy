package config

import (
	"errors"

	kmsv1 "github.com/Onurlulardan/harpy/contracts/kms/v1"
)

var errInvalidIAMKMSKeyID = errors.New("IAM KMS key ID is invalid")

type IAMKMSKeyID struct {
	value string
	valid bool
}

func loadIAMKMSKeyID(values environmentValues, key string) (IAMKMSKeyID, error) {
	value, exists := values.lookup(key)
	if !exists {
		return IAMKMSKeyID{}, configError(key, "is required")
	}
	keyID, ok := parseIAMKMSKeyID(value)
	if !ok {
		return IAMKMSKeyID{}, configError(key, "must satisfy iam-kms-key-id.v1")
	}
	return keyID, nil
}

func parseIAMKMSKeyID(value string) (IAMKMSKeyID, bool) {
	if _, err := kmsv1.ParseGoogleCloudKeyVersion(value); err != nil {
		return IAMKMSKeyID{}, false
	}
	return IAMKMSKeyID{value: value, valid: true}, true
}

func (keyID IAMKMSKeyID) Validate() error {
	parsed, ok := parseIAMKMSKeyID(keyID.value)
	if !keyID.valid || !ok || parsed.value != keyID.value {
		return errInvalidIAMKMSKeyID
	}
	return nil
}

func (keyID IAMKMSKeyID) Value() (string, error) {
	if err := keyID.Validate(); err != nil {
		return "", err
	}
	return keyID.value, nil
}

func (keyID IAMKMSKeyID) sameCryptoKey(other IAMKMSKeyID) bool {
	left, leftErr := kmsv1.ParseGoogleCloudKeyVersion(keyID.value)
	right, rightErr := kmsv1.ParseGoogleCloudKeyVersion(other.value)
	return keyID.valid && other.valid && leftErr == nil && rightErr == nil &&
		left.SameCryptoKey(right)
}
