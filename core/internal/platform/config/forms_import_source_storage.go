package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

const (
	formsImportSourceEndpointKey  = "HARPY_FORMS_IMPORT_SOURCE_S3_ENDPOINT"
	formsImportSourceRegionKey    = "HARPY_FORMS_IMPORT_SOURCE_S3_REGION"
	formsImportSourceBucketKey    = "HARPY_FORMS_IMPORT_SOURCE_S3_BUCKET"
	formsImportSourceOwnerKey     = "HARPY_FORMS_IMPORT_SOURCE_S3_EXPECTED_BUCKET_OWNER"
	formsImportSourcePathStyleKey = "HARPY_FORMS_IMPORT_SOURCE_S3_PATH_STYLE"
	formsImportSourceStagingKey   = "HARPY_FORMS_IMPORT_SOURCE_STAGING_DIR"
)

var errInvalidFormsImportSourceS3 = errors.New("Forms import source S3 config is invalid")

type FormsImportSourceS3 struct {
	endpoint       string
	region         string
	bucket         string
	expectedOwner  string
	pathStyle      bool
	allowPlainHTTP bool
	credentials    FormsImportSourceS3Credentials
	valid          bool
}

func loadFormsImportSourceS3(values environmentValues,
	environment Environment) (FormsImportSourceS3, error) {
	endpoint, err := requiredFormsImportSourceText(values, formsImportSourceEndpointKey, 2048)
	if err != nil || !validFormsExportEndpoint(endpoint, !environment.productionLike()) {
		return FormsImportSourceS3{}, configError(formsImportSourceEndpointKey,
			"must be a canonical permitted endpoint")
	}
	region, err := requiredFormsImportSourceText(values, formsImportSourceRegionKey, 63)
	if err != nil || !validFormsExportRegion(region) {
		return FormsImportSourceS3{}, configError(formsImportSourceRegionKey,
			"must be a canonical region")
	}
	bucket, err := requiredFormsImportSourceText(values, formsImportSourceBucketKey, 63)
	if err != nil || !validFormsExportBucket(bucket) {
		return FormsImportSourceS3{}, configError(formsImportSourceBucketKey,
			"must be a DNS-compatible bucket")
	}
	owner, err := requiredFormsImportSourceText(values, formsImportSourceOwnerKey, 12)
	if err != nil || len(owner) != 12 || !decimalDigits(owner) {
		return FormsImportSourceS3{}, configError(formsImportSourceOwnerKey,
			"must be a 12-digit owner ID")
	}
	if _, found := values.lookup(formsImportSourcePathStyleKey); !found {
		return FormsImportSourceS3{}, configError(formsImportSourcePathStyleKey, "is required")
	}
	pathStyle, err := parseBoolean(values, formsImportSourcePathStyleKey, "")
	if err != nil {
		return FormsImportSourceS3{}, err
	}
	credentials, err := loadFormsImportSourceS3Credentials(values)
	if err != nil {
		return FormsImportSourceS3{}, err
	}
	return FormsImportSourceS3{endpoint: endpoint, region: region, bucket: bucket,
		expectedOwner: owner, pathStyle: pathStyle,
		allowPlainHTTP: !environment.productionLike(), credentials: credentials,
		valid: true}, nil
}

func (settings FormsImportSourceS3) Validate() error {
	if !settings.valid || !validFormsExportEndpoint(settings.endpoint, settings.allowPlainHTTP) ||
		!validFormsExportRegion(settings.region) || !validFormsExportBucket(settings.bucket) ||
		len(settings.expectedOwner) != 12 || !decimalDigits(settings.expectedOwner) ||
		settings.credentials.Validate() != nil {
		return errInvalidFormsImportSourceS3
	}
	return nil
}

func (settings FormsImportSourceS3) Connection() (
	string, string, string, string, bool, error,
) {
	if err := settings.Validate(); err != nil {
		return "", "", "", "", false, err
	}
	return settings.endpoint, settings.region, settings.bucket,
		settings.expectedOwner, settings.pathStyle, nil
}

func (settings FormsImportSourceS3) Credentials() (FormsImportSourceS3Credentials, error) {
	if err := settings.Validate(); err != nil {
		return FormsImportSourceS3Credentials{}, err
	}
	return settings.credentials, nil
}

func (FormsImportSourceS3) String() string   { return "[REDACTED FORMS IMPORT SOURCE S3]" }
func (FormsImportSourceS3) GoString() string { return "[REDACTED FORMS IMPORT SOURCE S3]" }
func (FormsImportSourceS3) LogValue() slog.Value {
	return slog.StringValue("[REDACTED FORMS IMPORT SOURCE S3]")
}
func (FormsImportSourceS3) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED FORMS IMPORT SOURCE S3]")
}

func requiredFormsImportSourceText(values environmentValues, key string,
	maximum int) (string, error) {
	return requiredFormsExportText(values, key, maximum)
}
