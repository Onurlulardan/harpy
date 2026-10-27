package config

import (
	"errors"
	"strings"

	kmsv1 "github.com/Onurlulardan/harpy/contracts/kms/v1"
)

var errInvalidIAMKMSKeyID = errors.New("IAM KMS key ID is invalid")

type IAMKMSKeyID struct {
	value            string
	valid            bool
	localDevelopment bool
}

func loadIAMKMSKeyID(
	values environmentValues,
	key string,
	environment Environment,
) (IAMKMSKeyID, error) {
	value, exists := values.lookup(key)
	if !exists {
		if environment.productionLike() {
			return IAMKMSKeyID{}, configError(key, "is required")
		}
		return IAMKMSKeyID{
			value: "local-development:" + key, valid: true, localDevelopment: true,
		}, nil
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
	if keyID.localDevelopment {
		if keyID.valid && strings.HasPrefix(keyID.value, "local-development:") {
			return nil
		}
		return errInvalidIAMKMSKeyID
	}
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
