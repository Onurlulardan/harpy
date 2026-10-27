package environmentrecoveryv1

import (
	"encoding/json"
	"time"
)

type Purpose string

const (
	Schema                          = "harpy-recovery-artifact.v1"
	Algorithm                       = "Ed25519"
	MaximumArtifactBytes            = 2 * 1024 * 1024
	MaximumPayloadBytes             = 1024 * 1024
	PurposeRecoveryManifest Purpose = "recovery_manifest"
	PurposeRestorePlan      Purpose = "restore_plan"
	PurposeRestoreProgress  Purpose = "restore_progress"
	PurposeRestoreReceipt   Purpose = "restore_receipt"
)

type Input struct {
	Purpose       Purpose
	EnvironmentID string
	OperationID   string
	ArtifactID    string
	Sequence      int64
	IssuedAt      time.Time
	ExpiresAt     time.Time
	Payload       []byte
}

type wire struct {
	Schema        string          `json:"schema"`
	Purpose       Purpose         `json:"purpose"`
	EnvironmentID string          `json:"environment_id"`
	OperationID   string          `json:"operation_id"`
	ArtifactID    string          `json:"artifact_id"`
	Sequence      int64           `json:"artifact_sequence"`
	IssuedAt      string          `json:"issued_at"`
	ExpiresAt     string          `json:"expires_at"`
	PayloadSHA256 string          `json:"payload_sha256"`
	Payload       json.RawMessage `json:"payload"`
	KeyGeneration int64           `json:"key_generation"`
	SignerKID     string          `json:"signer_kid"`
	Algorithm     string          `json:"algorithm"`
	Signature     string          `json:"signature,omitempty"`
}

type Artifact struct {
	wire      wire
	canonical []byte
	digest    string
	issuedAt  time.Time
	expiresAt time.Time
}

func (value Artifact) Purpose() Purpose       { return value.wire.Purpose }
func (value Artifact) EnvironmentID() string  { return value.wire.EnvironmentID }
func (value Artifact) OperationID() string    { return value.wire.OperationID }
func (value Artifact) ArtifactID() string     { return value.wire.ArtifactID }
func (value Artifact) Sequence() int64        { return value.wire.Sequence }
func (value Artifact) KeyGeneration() int64   { return value.wire.KeyGeneration }
func (value Artifact) SignerKID() string      { return value.wire.SignerKID }
func (value Artifact) SHA256() string         { return value.digest }
func (value Artifact) IssuedAt() time.Time    { return value.issuedAt }
func (value Artifact) ExpiresAt() time.Time   { return value.expiresAt }
func (value Artifact) CanonicalBytes() []byte { return append([]byte(nil), value.canonical...) }
func (value Artifact) PayloadBytes() []byte   { return append([]byte(nil), value.wire.Payload...) }
