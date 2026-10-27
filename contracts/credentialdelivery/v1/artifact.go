package credentialdeliveryv1

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"time"

	"github.com/google/uuid"
)

const (
	ArtifactSchema        = "harpy-installation-credential-delivery-artifact.v1"
	InstallationAdapterID = "local-terminal"
	MaximumArtifactBytes  = 4096
)

var ErrInvalidArtifact = errors.New("credential delivery artifact is invalid")

type SealedArtifact struct {
	Schema           string `json:"schema"`
	OperationID      string `json:"operation_id"`
	AttemptID        string `json:"attempt_id"`
	AdapterID        string `json:"adapter_id"`
	ClaimTokenSHA256 string `json:"claim_token_sha256"`
	LeaseExpiresAt   string `json:"lease_expires_at"`
	Nonce            string `json:"nonce"`
	Ciphertext       string `json:"ciphertext"`
}

func ParseArtifact(document []byte) (SealedArtifact, error) {
	if len(document) == 0 || len(document) > MaximumArtifactBytes {
		return SealedArtifact{}, ErrInvalidArtifact
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var artifact SealedArtifact
	if err := decoder.Decode(&artifact); err != nil {
		return SealedArtifact{}, ErrInvalidArtifact
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return SealedArtifact{}, ErrInvalidArtifact
	}
	if !artifact.valid() {
		return SealedArtifact{}, ErrInvalidArtifact
	}
	return artifact, nil
}

func (artifact SealedArtifact) valid() bool {
	operationID, operationErr := uuid.Parse(artifact.OperationID)
	attemptID, attemptErr := uuid.Parse(artifact.AttemptID)
	lease, leaseErr := time.Parse(time.RFC3339Nano, artifact.LeaseExpiresAt)
	nonce, nonceErr := base64.RawURLEncoding.Strict().DecodeString(artifact.Nonce)
	ciphertext, ciphertextErr := base64.RawURLEncoding.Strict().DecodeString(artifact.Ciphertext)
	return artifact.Schema == ArtifactSchema && operationErr == nil &&
		operationID.Version() == uuid.Version(7) && attemptErr == nil && attemptID != uuid.Nil &&
		artifact.AdapterID == InstallationAdapterID && digestPattern.MatchString(artifact.ClaimTokenSHA256) &&
		leaseErr == nil && lease.Location() == time.UTC && nonceErr == nil && len(nonce) == 12 &&
		ciphertextErr == nil && len(ciphertext) == 48
}

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func artifactAAD(artifact SealedArtifact) []byte {
	return []byte(artifact.Schema + "\x00" + artifact.OperationID + "\x00" +
		artifact.AttemptID + "\x00" + artifact.AdapterID + "\x00" +
		artifact.ClaimTokenSHA256 + "\x00" + artifact.LeaseExpiresAt)
}
