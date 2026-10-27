package credentialdeliveryv1

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const artifactKeyDomain = "harpy-installation-credential-delivery-key.v1\x00"

func SealArtifact(
	operationID, attemptID uuid.UUID,
	claimTokenSHA256 string,
	leaseExpiresAt time.Time,
	credential, plaintext []byte,
) ([]byte, error) {
	if operationID.Version() != uuid.Version(7) || attemptID == uuid.Nil ||
		!digestPattern.MatchString(claimTokenSHA256) || len(plaintext) != 32 ||
		leaseExpiresAt.IsZero() || leaseExpiresAt.Location() != time.UTC {
		return nil, ErrInvalidArtifact
	}
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return sealArtifactWithNonce(operationID, attemptID, claimTokenSHA256,
		leaseExpiresAt, credential, plaintext, nonce)
}

func sealArtifactWithNonce(
	operationID, attemptID uuid.UUID,
	claimTokenSHA256 string,
	leaseExpiresAt time.Time,
	credential, plaintext, nonce []byte,
) ([]byte, error) {
	if operationID.Version() != uuid.Version(7) || attemptID == uuid.Nil ||
		!digestPattern.MatchString(claimTokenSHA256) || len(plaintext) != 32 ||
		leaseExpiresAt.IsZero() || leaseExpiresAt.Location() != time.UTC || len(nonce) != 12 {
		return nil, ErrInvalidArtifact
	}
	key, err := artifactKey(operationID, credential)
	if err != nil {
		return nil, err
	}
	defer clear(key)
	aead, err := artifactAEAD(key)
	if err != nil {
		return nil, err
	}
	artifact := SealedArtifact{
		Schema: ArtifactSchema, OperationID: operationID.String(), AttemptID: attemptID.String(),
		AdapterID: InstallationAdapterID, ClaimTokenSHA256: claimTokenSHA256,
		LeaseExpiresAt: leaseExpiresAt.Format(time.RFC3339Nano),
		Nonce:          base64.RawURLEncoding.EncodeToString(nonce),
	}
	sealed := aead.Seal(nil, nonce, plaintext, artifactAAD(artifact))
	artifact.Ciphertext = base64.RawURLEncoding.EncodeToString(sealed)
	return json.Marshal(artifact)
}

func OpenArtifact(
	document []byte,
	expectedOperationID uuid.UUID,
	credential []byte,
) (SealedArtifact, []byte, error) {
	artifact, err := ParseArtifact(document)
	if err != nil || artifact.OperationID != expectedOperationID.String() {
		return SealedArtifact{}, nil, ErrInvalidArtifact
	}
	key, err := artifactKey(expectedOperationID, credential)
	if err != nil {
		return SealedArtifact{}, nil, err
	}
	defer clear(key)
	aead, err := artifactAEAD(key)
	if err != nil {
		return SealedArtifact{}, nil, err
	}
	nonce, _ := base64.RawURLEncoding.Strict().DecodeString(artifact.Nonce)
	ciphertext, _ := base64.RawURLEncoding.Strict().DecodeString(artifact.Ciphertext)
	plaintext, err := aead.Open(nil, nonce, ciphertext, artifactAAD(artifact))
	if err != nil || len(plaintext) != 32 {
		clear(plaintext)
		return SealedArtifact{}, nil, ErrInvalidArtifact
	}
	return artifact, plaintext, nil
}

func artifactKey(operationID uuid.UUID, credential []byte) ([]byte, error) {
	raw, err := base64.RawURLEncoding.Strict().DecodeString(string(credential))
	if err != nil || len(raw) != 32 {
		clear(raw)
		return nil, ErrInvalidArtifact
	}
	defer clear(raw)
	mac := hmac.New(sha256.New, raw)
	_, _ = mac.Write([]byte(artifactKeyDomain + operationID.String()))
	return mac.Sum(nil), nil
}

func artifactAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
