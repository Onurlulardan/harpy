package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"strings"
)

const (
	formsExportEndpointKey  = "HARPY_FORMS_EXPORT_S3_ENDPOINT"
	formsExportRegionKey    = "HARPY_FORMS_EXPORT_S3_REGION"
	formsExportBucketKey    = "HARPY_FORMS_EXPORT_S3_BUCKET"
	formsExportOwnerKey     = "HARPY_FORMS_EXPORT_S3_EXPECTED_BUCKET_OWNER"
	formsExportPathStyleKey = "HARPY_FORMS_EXPORT_S3_PATH_STYLE"
)

var errInvalidFormsExportS3 = errors.New("Forms export S3 config is invalid")

type FormsExportS3 struct {
	endpoint       string
	region         string
	bucket         string
	expectedOwner  string
	pathStyle      bool
	ownerOptional  bool
	allowPlainHTTP bool
	credentials    FormsExportS3Credentials
	valid          bool
}

func loadFormsExportS3(
	values environmentValues,
	environment Environment,
) (FormsExportS3, error) {
	endpoint, err := requiredFormsExportText(values, formsExportEndpointKey, 2048)
	if err != nil || !validFormsExportEndpoint(endpoint, !environment.productionLike()) {
		return FormsExportS3{}, configError(formsExportEndpointKey, "must be a canonical permitted endpoint")
	}
	region, err := requiredFormsExportText(values, formsExportRegionKey, 63)
	if err != nil || !validFormsExportRegion(region) {
		return FormsExportS3{}, configError(formsExportRegionKey, "must be a canonical region")
	}
	bucket, err := requiredFormsExportText(values, formsExportBucketKey, 63)
	if err != nil || !validFormsExportBucket(bucket) {
		return FormsExportS3{}, configError(formsExportBucketKey, "must be a DNS-compatible bucket")
	}
	if _, found := values.lookup(formsExportPathStyleKey); !found {
		return FormsExportS3{}, configError(formsExportPathStyleKey, "is required")
	}
	pathStyle, err := parseBoolean(values, formsExportPathStyleKey, "")
	if err != nil {
		return FormsExportS3{}, err
	}
	owner, ownerOptional, err := resolveBucketOwner(values, formsExportOwnerKey, environment, pathStyle)
	if err != nil {
		return FormsExportS3{}, err
	}
	credentials, err := loadFormsExportS3Credentials(values)
	if err != nil {
		return FormsExportS3{}, err
	}
	return FormsExportS3{
		endpoint: endpoint, region: region, bucket: bucket, expectedOwner: owner,
		pathStyle: pathStyle, ownerOptional: ownerOptional,
		allowPlainHTTP: !environment.productionLike(),
		credentials:    credentials, valid: true,
	}, nil
}

func (settings FormsExportS3) Validate() error {
	if !settings.valid || !validFormsExportEndpoint(settings.endpoint, settings.allowPlainHTTP) ||
		!validFormsExportRegion(settings.region) || !validFormsExportBucket(settings.bucket) ||
		!bucketOwnerAccepted(settings.expectedOwner, settings.ownerOptional) ||
		settings.credentials.Validate() != nil {
		return errInvalidFormsExportS3
	}
	return nil
}

func (settings FormsExportS3) Connection() (string, string, string, string, bool, error) {
	if err := settings.Validate(); err != nil {
		return "", "", "", "", false, err
	}
	return settings.endpoint, settings.region, settings.bucket,
		settings.expectedOwner, settings.pathStyle, nil
}

func (settings FormsExportS3) Credentials() (FormsExportS3Credentials, error) {
	if err := settings.Validate(); err != nil {
		return FormsExportS3Credentials{}, err
	}
	return settings.credentials, nil
}

func (FormsExportS3) String() string   { return "[REDACTED FORMS EXPORT S3 CONFIG]" }
func (FormsExportS3) GoString() string { return "[REDACTED FORMS EXPORT S3 CONFIG]" }
func (FormsExportS3) LogValue() slog.Value {
	return slog.StringValue("[REDACTED FORMS EXPORT S3 CONFIG]")
}
func (FormsExportS3) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED FORMS EXPORT S3 CONFIG]")
}

func requiredFormsExportText(values environmentValues, key string, maximum int) (string, error) {
	value, found := values.lookup(key)
	if !found || value == "" || len(value) > maximum || strings.TrimSpace(value) != value ||
		strings.ContainsAny(value, "\x00\r\n\t") {
		return "", configError(key, "is required and must be canonical")
	}
	return value, nil
}

func validFormsExportEndpoint(value string, allowHTTP bool) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed == nil || parsed.Host == "" || parsed.User != nil ||
		parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" ||
		parsed.String() != value || parsed.Host != strings.ToLower(parsed.Host) {
		return false
	}
	return parsed.Scheme == "https" || allowHTTP && parsed.Scheme == "http"
}

func validFormsExportRegion(value string) bool {
	if len(value) < 1 || len(value) > 63 || !lowercaseLetterOrDigit(value[0]) {
		return false
	}
	for index := 1; index < len(value); index++ {
		if !lowercaseLetterOrDigit(value[index]) && value[index] != '-' {
			return false
		}
	}
	return true
}

func validFormsExportBucket(value string) bool {
	if len(value) < 3 || len(value) > 63 || net.ParseIP(value) != nil ||
		strings.Contains(value, "..") {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if label == "" || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if character != '-' && (character < 'a' || character > 'z') &&
				(character < '0' || character > '9') {
				return false
			}
		}
	}
	return true
}
