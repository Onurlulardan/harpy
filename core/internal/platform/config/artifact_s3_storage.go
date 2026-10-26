package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

var errInvalidArtifactS3 = errors.New("artifact S3 config is invalid")

type artifactStorageKeys struct {
	endpoint, region, bucket, owner, pathStyle string
	credentials                                artifactCredentialKeys
}

type ArtifactS3 struct {
	endpoint       string
	region         string
	bucket         string
	expectedOwner  string
	pathStyle      bool
	allowPlainHTTP bool
	credentials    ArtifactS3Credentials
	valid          bool
}

func loadArtifactS3(values environmentValues, environment Environment,
	keys artifactStorageKeys) (ArtifactS3, error) {
	endpoint, err := requiredFormsExportText(values, keys.endpoint, 2048)
	if err != nil || !validFormsExportEndpoint(endpoint, !environment.productionLike()) {
		return ArtifactS3{}, configError(keys.endpoint,
			"must be a canonical permitted endpoint")
	}
	region, err := requiredFormsExportText(values, keys.region, 63)
	if err != nil || !validFormsExportRegion(region) {
		return ArtifactS3{}, configError(keys.region, "must be a canonical region")
	}
	bucket, err := requiredFormsExportText(values, keys.bucket, 63)
	if err != nil || !validFormsExportBucket(bucket) {
		return ArtifactS3{}, configError(keys.bucket, "must be a DNS-compatible bucket")
	}
	owner, err := requiredFormsExportText(values, keys.owner, 12)
	if err != nil || len(owner) != 12 || !decimalDigits(owner) {
		return ArtifactS3{}, configError(keys.owner, "must be a 12-digit owner ID")
	}
	if _, found := values.lookup(keys.pathStyle); !found {
		return ArtifactS3{}, configError(keys.pathStyle, "is required")
	}
	pathStyle, err := parseBoolean(values, keys.pathStyle, "")
	if err != nil {
		return ArtifactS3{}, err
	}
	credentials, err := loadArtifactS3Credentials(values, keys.credentials)
	if err != nil {
		return ArtifactS3{}, err
	}
	return ArtifactS3{endpoint: endpoint, region: region, bucket: bucket,
		expectedOwner: owner, pathStyle: pathStyle,
		allowPlainHTTP: !environment.productionLike(), credentials: credentials,
		valid: true}, nil
}

func (settings ArtifactS3) Validate() error {
	if !settings.valid ||
		!validFormsExportEndpoint(settings.endpoint, settings.allowPlainHTTP) ||
		!validFormsExportRegion(settings.region) ||
		!validFormsExportBucket(settings.bucket) || len(settings.expectedOwner) != 12 ||
		!decimalDigits(settings.expectedOwner) || settings.credentials.Validate() != nil {
		return errInvalidArtifactS3
	}
	return nil
}

func (settings ArtifactS3) Connection() (string, string, string, string, bool, error) {
	if err := settings.Validate(); err != nil {
		return "", "", "", "", false, err
	}
	return settings.endpoint, settings.region, settings.bucket,
		settings.expectedOwner, settings.pathStyle, nil
}

func (settings ArtifactS3) Credentials() (ArtifactS3Credentials, error) {
	if err := settings.Validate(); err != nil {
		return ArtifactS3Credentials{}, err
	}
	return settings.credentials, nil
}

func (ArtifactS3) String() string   { return "[REDACTED ARTIFACT S3 CONFIG]" }
func (ArtifactS3) GoString() string { return "[REDACTED ARTIFACT S3 CONFIG]" }
func (ArtifactS3) LogValue() slog.Value {
	return slog.StringValue("[REDACTED ARTIFACT S3 CONFIG]")
}
func (ArtifactS3) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED ARTIFACT S3 CONFIG]")
}
