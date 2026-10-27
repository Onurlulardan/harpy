package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

type ArtifactCredentialSource string

const (
	ArtifactWorkloadIdentity ArtifactCredentialSource = "workload_identity"
	ArtifactStaticFiles      ArtifactCredentialSource = "static_files"
)

var errInvalidArtifactCredentials = errors.New("artifact S3 credentials are invalid")

type artifactCredentialKeys struct {
	source, access, secret, session string
}

type ArtifactS3Credentials struct {
	source       ArtifactCredentialSource
	accessKeyID  Secret
	secretKey    Secret
	sessionToken Secret
	hasSession   bool
	valid        bool
}

func loadArtifactS3Credentials(values environmentValues,
	keys artifactCredentialKeys) (ArtifactS3Credentials, error) {
	raw, found := values.lookup(keys.source)
	if !found {
		return ArtifactS3Credentials{}, configError(keys.source, "is required")
	}
	source := ArtifactCredentialSource(raw)
	staticConfigured := artifactStaticFilesConfigured(values, keys)
	if source == ArtifactWorkloadIdentity {
		if staticConfigured {
			return ArtifactS3Credentials{}, configError(keys.source,
				"workload identity forbids static credential files")
		}
		return ArtifactS3Credentials{source: source, valid: true}, nil
	}
	if source != ArtifactStaticFiles {
		return ArtifactS3Credentials{}, configError(keys.source,
			"must be workload_identity or static_files")
	}
	access, err := loadArtifactCredentialFile(values, keys.access, 128)
	if err != nil {
		return ArtifactS3Credentials{}, err
	}
	secret, err := loadArtifactCredentialFile(values, keys.secret, 1024)
	if err != nil {
		return ArtifactS3Credentials{}, err
	}
	token, hasToken, err := loadOptionalArtifactCredentialFile(values, keys.session, 4096)
	if err != nil {
		return ArtifactS3Credentials{}, err
	}
	return ArtifactS3Credentials{source: source, accessKeyID: access,
		secretKey: secret, sessionToken: token, hasSession: hasToken, valid: true}, nil
}

func (credentials ArtifactS3Credentials) Validate() error {
	if !credentials.valid {
		return errInvalidArtifactCredentials
	}
	if credentials.source == ArtifactWorkloadIdentity {
		if credentials.accessKeyID.Value() != "" || credentials.secretKey.Value() != "" ||
			credentials.sessionToken.Value() != "" || credentials.hasSession {
			return errInvalidArtifactCredentials
		}
		return nil
	}
	if credentials.source != ArtifactStaticFiles ||
		!validArtifactCredential(credentials.accessKeyID.Value()) ||
		!validArtifactCredential(credentials.secretKey.Value()) ||
		credentials.hasSession != (credentials.sessionToken.Value() != "") {
		return errInvalidArtifactCredentials
	}
	return nil
}

func (credentials ArtifactS3Credentials) Source() (ArtifactCredentialSource, error) {
	if err := credentials.Validate(); err != nil {
		return "", err
	}
	return credentials.source, nil
}

func (credentials ArtifactS3Credentials) Static() (Secret, Secret, Secret, bool, error) {
	if credentials.Validate() != nil || credentials.source != ArtifactStaticFiles {
		return Secret{}, Secret{}, Secret{}, false, errInvalidArtifactCredentials
	}
	return credentials.accessKeyID, credentials.secretKey,
		credentials.sessionToken, credentials.hasSession, nil
}

func artifactStaticFilesConfigured(values environmentValues, keys artifactCredentialKeys) bool {
	for _, key := range []string{keys.access, keys.secret, keys.session} {
		if _, found := values.lookup(key); found {
			return true
		}
	}
	return false
}

func loadArtifactCredentialFile(values environmentValues, key string,
	maximum int) (Secret, error) {
	value, found, err := loadOptionalArtifactCredentialFile(values, key, maximum)
	if err != nil {
		return Secret{}, err
	}
	if !found {
		return Secret{}, configError(key, "is required for static_files")
	}
	return value, nil
}

func loadOptionalArtifactCredentialFile(values environmentValues, key string,
	maximum int) (Secret, bool, error) {
	path, found := values.lookup(key)
	if !found {
		return Secret{}, false, nil
	}
	value, err := readSecretFile(key, path, maximum)
	if err != nil || !validArtifactCredential(value) {
		return Secret{}, false, configError(key, "is invalid")
	}
	return Secret{value: value}, true, nil
}

func (ArtifactS3Credentials) String() string   { return "[REDACTED]" }
func (ArtifactS3Credentials) GoString() string { return "[REDACTED]" }
func (ArtifactS3Credentials) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}
func (ArtifactS3Credentials) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED]")
}
