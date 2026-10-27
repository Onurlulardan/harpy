package recoverykitv1

import (
	"crypto/ed25519"
	"errors"
	"time"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

var ErrMaterial = errors.New("environment recovery kit material is invalid")

const (
	PlaintextSchema = "environment-recovery-kit.v1"
	SealedSchema    = "environment-recovery-kit.hpke.v1"
	Suite           = "DHKEM(X25519,HKDF-SHA256)/HKDF-SHA256/AES-256-GCM"
	MaximumKitBytes = 2 * 1024 * 1024
)

type RecoveryKeyMaterial struct {
	KeyID      string
	PrivateKey [32]byte
	State      string
}

type SigningKeyMaterial struct {
	Purpose    recoverysigningv1.Purpose
	KeyID      string
	PrivateKey ed25519.PrivateKey
}

type BuildInput struct {
	EnvironmentID, OperationID string
	Generation                 int64
	DeploymentBindingSHA256    [32]byte
	RecoveryKeys               []RecoveryKeyMaterial
	SigningKeys                []SigningKeyMaterial
	TrustBundle                recoverysigningv1.Bundle
	AckProof                   [32]byte
	PreviousKitSHA256          *[32]byte
	GeneratedAt                time.Time
}

type SealInput struct {
	EnvironmentID, OperationID string
	Generation                 int64
	ClaimPublicKey             [32]byte
	IssuedAt, ExpiresAt        time.Time
}

type OpenExpected struct {
	EnvironmentID, OperationID string
	Generation                 int64
	Now                        time.Time
}

type recoveryKeyWire struct {
	KeyID      string `json:"key_id"`
	Algorithm  string `json:"algorithm"`
	PrivateKey string `json:"private_key_base64url"`
	State      string `json:"state"`
}

type signingKeyWire struct {
	Purpose    recoverysigningv1.Purpose `json:"purpose"`
	KeyID      string                    `json:"key_id"`
	Algorithm  string                    `json:"algorithm"`
	PrivateKey string                    `json:"private_key_base64url"`
}

type plaintextWire struct {
	Schema                       string            `json:"schema"`
	EnvironmentID                string            `json:"environment_id"`
	OperationID                  string            `json:"operation_id"`
	Generation                   int64             `json:"generation"`
	DeploymentBindingSHA256      string            `json:"deployment_binding_sha256"`
	RecoveryKeys                 []recoveryKeyWire `json:"recovery_keys"`
	SigningKeys                  []signingKeyWire  `json:"signing_keys"`
	RecoveryTrustBundleBase64URL string            `json:"recovery_trust_bundle_base64url"`
	AckProofBase64URL            string            `json:"ack_proof_base64url"`
	PreviousKitSHA256            *string           `json:"previous_kit_sha256"`
	GeneratedAt                  string            `json:"generated_at"`
}

type sealedMetadata struct {
	Schema               string `json:"schema"`
	Suite                string `json:"suite"`
	EnvironmentID        string `json:"environment_id"`
	OperationID          string `json:"operation_id"`
	Generation           int64  `json:"generation"`
	ClaimPublicKeySHA256 string `json:"claim_public_key_sha256"`
	IssuedAt             string `json:"issued_at"`
	ExpiresAt            string `json:"expires_at"`
}

type sealedWire struct {
	Schema               string `json:"schema"`
	Suite                string `json:"suite"`
	EnvironmentID        string `json:"environment_id"`
	OperationID          string `json:"operation_id"`
	Generation           int64  `json:"generation"`
	ClaimPublicKeySHA256 string `json:"claim_public_key_sha256"`
	IssuedAt             string `json:"issued_at"`
	ExpiresAt            string `json:"expires_at"`
	EncapsulatedKey      string `json:"encapsulated_key_base64url"`
	Ciphertext           string `json:"ciphertext_base64url"`
}

type Kit struct {
	wire      plaintextWire
	canonical []byte
}

func (kit Kit) CanonicalBytes() []byte    { return append([]byte(nil), kit.canonical...) }
func (kit Kit) EnvironmentID() string     { return kit.wire.EnvironmentID }
func (kit Kit) OperationID() string       { return kit.wire.OperationID }
func (kit Kit) Generation() int64         { return kit.wire.Generation }
func (kit Kit) AckProofBase64URL() string { return kit.wire.AckProofBase64URL }
