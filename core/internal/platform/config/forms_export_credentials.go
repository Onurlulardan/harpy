package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

const (
	formsExportCredentialSourceKey = "HARPY_FORMS_EXPORT_S3_CREDENTIAL_SOURCE"
	formsExportAccessKeyFileKey    = "HARPY_FORMS_EXPORT_S3_ACCESS_KEY_ID_FILE"
	formsExportSecretKeyFileKey    = "HARPY_FORMS_EXPORT_S3_SECRET_ACCESS_KEY_FILE"
	formsExportSessionTokenFileKey = "HARPY_FORMS_EXPORT_S3_SESSION_TOKEN_FILE"
)

type FormsExportCredentialSource string

const (
	FormsExportWorkloadIdentity FormsExportCredentialSource = "workload_identity"
	FormsExportStaticFiles      FormsExportCredentialSource = "static_files"
)

var errInvalidFormsExportCredentials = errors.New("Forms export S3 credentials are invalid")

type FormsExportS3Credentials struct {
	source       FormsExportCredentialSource
	accessKeyID  Secret
	secretKey    Secret
	sessionToken Secret
	hasSession   bool
	valid        bool
}

func loadFormsExportS3Credentials(values environmentValues) (FormsExportS3Credentials, error) {
	raw, found := values.lookup(formsExportCredentialSourceKey)
	if !found {
		return FormsExportS3Credentials{}, configError(formsExportCredentialSourceKey, "is required")
	}
	source := FormsExportCredentialSource(raw)
	staticConfigured := formsExportStaticFilesConfigured(values)
	if source == FormsExportWorkloadIdentity {
		if staticConfigured {
			return FormsExportS3Credentials{}, configError(formsExportCredentialSourceKey, "workload identity forbids static credential files")
		}
		return FormsExportS3Credentials{source: source, valid: true}, nil
	}
	if source != FormsExportStaticFiles {
		return FormsExportS3Credentials{}, configError(formsExportCredentialSourceKey, "must be workload_identity or static_files")
	}
	access, err := loadFormsExportCredentialFile(values, formsExportAccessKeyFileKey, 128)
	if err != nil {
		return FormsExportS3Credentials{}, err
	}
	secret, err := loadFormsExportCredentialFile(values, formsExportSecretKeyFileKey, 1024)
	if err != nil {
		return FormsExportS3Credentials{}, err
	}
	token, hasToken, err := loadOptionalFormsExportCredentialFile(values, formsExportSessionTokenFileKey, 4096)
	if err != nil {
		return FormsExportS3Credentials{}, err
	}
	return FormsExportS3Credentials{
		source: source, accessKeyID: access, secretKey: secret,
		sessionToken: token, hasSession: hasToken, valid: true,
	}, nil
}

func (credentials FormsExportS3Credentials) Validate() error {
	if !credentials.valid {
		return errInvalidFormsExportCredentials
	}
	if credentials.source == FormsExportWorkloadIdentity {
		if credentials.accessKeyID.Value() != "" || credentials.secretKey.Value() != "" ||
			credentials.sessionToken.Value() != "" || credentials.hasSession {
			return errInvalidFormsExportCredentials
		}
		return nil
	}
	if credentials.source != FormsExportStaticFiles ||
		!validFormsExportCredential(credentials.accessKeyID.Value()) ||
		!validFormsExportCredential(credentials.secretKey.Value()) ||
		credentials.hasSession != (credentials.sessionToken.Value() != "") {
		return errInvalidFormsExportCredentials
	}
	return nil
}

func (credentials FormsExportS3Credentials) Source() (FormsExportCredentialSource, error) {
	if err := credentials.Validate(); err != nil {
		return "", err
	}
	return credentials.source, nil
}

func (credentials FormsExportS3Credentials) Static() (Secret, Secret, Secret, bool, error) {
	if err := credentials.Validate(); err != nil || credentials.source != FormsExportStaticFiles {
		return Secret{}, Secret{}, Secret{}, false, errInvalidFormsExportCredentials
	}
	return credentials.accessKeyID, credentials.secretKey,
		credentials.sessionToken, credentials.hasSession, nil
}

func (FormsExportS3Credentials) String() string   { return "[REDACTED]" }
func (FormsExportS3Credentials) GoString() string { return "[REDACTED]" }
func (FormsExportS3Credentials) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}
func (FormsExportS3Credentials) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED]")
}

func formsExportStaticFilesConfigured(values environmentValues) bool {
	for _, key := range []string{
		formsExportAccessKeyFileKey, formsExportSecretKeyFileKey, formsExportSessionTokenFileKey,
	} {
		if _, found := values.lookup(key); found {
			return true
		}
	}
	return false
}

func loadFormsExportCredentialFile(
	values environmentValues, key string, maximum int,
) (Secret, error) {
	value, found, err := loadOptionalFormsExportCredentialFile(values, key, maximum)
	if err != nil {
		return Secret{}, err
	}
	if !found {
		return Secret{}, configError(key, "is required for static_files")
	}
	return value, nil
}

func loadOptionalFormsExportCredentialFile(
	values environmentValues, key string, maximum int,
) (Secret, bool, error) {
	path, found := values.lookup(key)
	if !found {
		return Secret{}, false, nil
	}
	value, err := readSecretFile(key, path, maximum)
	if err != nil || !validFormsExportCredential(value) {
		return Secret{}, false, configError(key, "is invalid")
	}
	return Secret{value: value}, true, nil
}

func validFormsExportCredential(value string) bool {
	if value == "" {
		return false
	}
	for index := 0; index < len(value); index++ {
		if value[index] < 0x21 || value[index] > 0x7e {
			return false
		}
	}
	return true
}
