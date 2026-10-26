package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

const (
	formsImportSourceCredentialSourceKey = "HARPY_FORMS_IMPORT_SOURCE_S3_CREDENTIAL_SOURCE"
	formsImportSourceAccessKeyFileKey    = "HARPY_FORMS_IMPORT_SOURCE_S3_ACCESS_KEY_ID_FILE"
	formsImportSourceSecretKeyFileKey    = "HARPY_FORMS_IMPORT_SOURCE_S3_SECRET_ACCESS_KEY_FILE"
	formsImportSourceSessionTokenFileKey = "HARPY_FORMS_IMPORT_SOURCE_S3_SESSION_TOKEN_FILE"
)

type FormsImportSourceCredentialSource string

const (
	FormsImportSourceWorkloadIdentity FormsImportSourceCredentialSource = "workload_identity"
	FormsImportSourceStaticFiles      FormsImportSourceCredentialSource = "static_files"
)

var errInvalidFormsImportSourceCredentials = errors.New(
	"Forms import source S3 credentials are invalid")

type FormsImportSourceS3Credentials struct {
	source       FormsImportSourceCredentialSource
	accessKeyID  Secret
	secretKey    Secret
	sessionToken Secret
	hasSession   bool
	valid        bool
}

func loadFormsImportSourceS3Credentials(values environmentValues) (
	FormsImportSourceS3Credentials, error,
) {
	raw, found := values.lookup(formsImportSourceCredentialSourceKey)
	if !found {
		return FormsImportSourceS3Credentials{}, configError(
			formsImportSourceCredentialSourceKey, "is required")
	}
	source := FormsImportSourceCredentialSource(raw)
	staticConfigured := formsImportSourceStaticFilesConfigured(values)
	if source == FormsImportSourceWorkloadIdentity {
		if staticConfigured {
			return FormsImportSourceS3Credentials{}, configError(
				formsImportSourceCredentialSourceKey,
				"workload identity forbids static credential files")
		}
		return FormsImportSourceS3Credentials{source: source, valid: true}, nil
	}
	if source != FormsImportSourceStaticFiles {
		return FormsImportSourceS3Credentials{}, configError(
			formsImportSourceCredentialSourceKey, "must be workload_identity or static_files")
	}
	access, err := loadFormsImportSourceCredentialFile(
		values, formsImportSourceAccessKeyFileKey, 128)
	if err != nil {
		return FormsImportSourceS3Credentials{}, err
	}
	secret, err := loadFormsImportSourceCredentialFile(
		values, formsImportSourceSecretKeyFileKey, 1024)
	if err != nil {
		return FormsImportSourceS3Credentials{}, err
	}
	token, hasToken, err := loadOptionalFormsImportSourceCredentialFile(
		values, formsImportSourceSessionTokenFileKey, 4096)
	if err != nil {
		return FormsImportSourceS3Credentials{}, err
	}
	return FormsImportSourceS3Credentials{source: source, accessKeyID: access,
		secretKey: secret, sessionToken: token, hasSession: hasToken, valid: true}, nil
}

func (credentials FormsImportSourceS3Credentials) Validate() error {
	if !credentials.valid {
		return errInvalidFormsImportSourceCredentials
	}
	if credentials.source == FormsImportSourceWorkloadIdentity {
		if credentials.accessKeyID.Value() != "" || credentials.secretKey.Value() != "" ||
			credentials.sessionToken.Value() != "" || credentials.hasSession {
			return errInvalidFormsImportSourceCredentials
		}
		return nil
	}
	if credentials.source != FormsImportSourceStaticFiles ||
		!validFormsExportCredential(credentials.accessKeyID.Value()) ||
		!validFormsExportCredential(credentials.secretKey.Value()) ||
		credentials.hasSession != (credentials.sessionToken.Value() != "") {
		return errInvalidFormsImportSourceCredentials
	}
	return nil
}

func (credentials FormsImportSourceS3Credentials) Source() (
	FormsImportSourceCredentialSource, error,
) {
	if err := credentials.Validate(); err != nil {
		return "", err
	}
	return credentials.source, nil
}

func (credentials FormsImportSourceS3Credentials) Static() (
	Secret, Secret, Secret, bool, error,
) {
	if credentials.Validate() != nil || credentials.source != FormsImportSourceStaticFiles {
		return Secret{}, Secret{}, Secret{}, false, errInvalidFormsImportSourceCredentials
	}
	return credentials.accessKeyID, credentials.secretKey,
		credentials.sessionToken, credentials.hasSession, nil
}

func (FormsImportSourceS3Credentials) String() string   { return "[REDACTED]" }
func (FormsImportSourceS3Credentials) GoString() string { return "[REDACTED]" }
func (FormsImportSourceS3Credentials) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}
func (FormsImportSourceS3Credentials) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED]")
}

func formsImportSourceStaticFilesConfigured(values environmentValues) bool {
	for _, key := range []string{formsImportSourceAccessKeyFileKey,
		formsImportSourceSecretKeyFileKey, formsImportSourceSessionTokenFileKey} {
		if _, found := values.lookup(key); found {
			return true
		}
	}
	return false
}

func loadFormsImportSourceCredentialFile(values environmentValues,
	key string, maximum int) (Secret, error) {
	value, found, err := loadOptionalFormsImportSourceCredentialFile(values, key, maximum)
	if err != nil {
		return Secret{}, err
	}
	if !found {
		return Secret{}, configError(key, "is required for static_files")
	}
	return value, nil
}

func loadOptionalFormsImportSourceCredentialFile(values environmentValues,
	key string, maximum int) (Secret, bool, error) {
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
