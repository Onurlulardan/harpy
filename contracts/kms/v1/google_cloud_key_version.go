package kmsv1

import (
	"errors"
	"regexp"
	"strings"
)

const GoogleCloudKeyVersionPrefix = "gcp-kms://"

var (
	ErrGoogleCloudKeyVersion = errors.New("Google Cloud KMS key version is invalid")
	projectSegment           = regexp.MustCompile(`^(?:[a-z][a-z0-9-]{4,28}[a-z0-9]|[0-9]{6,30})$`)
	locationSegment          = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	keySegment               = regexp.MustCompile(`^[A-Za-z0-9_-]{1,63}$`)
	versionSegment           = regexp.MustCompile(`^[1-9][0-9]{0,18}$`)
)

type GoogleCloudKeyVersion struct {
	value     string
	resource  string
	cryptoKey string
}

func ParseGoogleCloudKeyVersion(value string) (GoogleCloudKeyVersion, error) {
	if !strings.HasPrefix(value, GoogleCloudKeyVersionPrefix) || len(value) > 512 {
		return GoogleCloudKeyVersion{}, ErrGoogleCloudKeyVersion
	}
	resource := strings.TrimPrefix(value, GoogleCloudKeyVersionPrefix)
	parts := strings.Split(resource, "/")
	if len(parts) != 10 || parts[0] != "projects" ||
		parts[2] != "locations" || parts[4] != "keyRings" ||
		parts[6] != "cryptoKeys" || parts[8] != "cryptoKeyVersions" ||
		!projectSegment.MatchString(parts[1]) ||
		!locationSegment.MatchString(parts[3]) ||
		!keySegment.MatchString(parts[5]) || !keySegment.MatchString(parts[7]) ||
		!versionSegment.MatchString(parts[9]) {
		return GoogleCloudKeyVersion{}, ErrGoogleCloudKeyVersion
	}
	return GoogleCloudKeyVersion{
		value: value, resource: resource, cryptoKey: strings.Join(parts[:8], "/"),
	}, nil
}

func (version GoogleCloudKeyVersion) Validate() error {
	parsed, err := ParseGoogleCloudKeyVersion(version.value)
	if err != nil || parsed.resource != version.resource ||
		parsed.cryptoKey != version.cryptoKey {
		return ErrGoogleCloudKeyVersion
	}
	return nil
}

func (version GoogleCloudKeyVersion) Value() (string, error) {
	if err := version.Validate(); err != nil {
		return "", err
	}
	return version.value, nil
}

func (version GoogleCloudKeyVersion) Resource() (string, error) {
	if err := version.Validate(); err != nil {
		return "", err
	}
	return version.resource, nil
}

func (version GoogleCloudKeyVersion) SameCryptoKey(other GoogleCloudKeyVersion) bool {
	return version.Validate() == nil && other.Validate() == nil &&
		version.cryptoKey == other.cryptoKey
}
