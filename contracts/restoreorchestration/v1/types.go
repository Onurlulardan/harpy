package restoreorchestrationv1

import (
	"crypto/ed25519"
	"time"
)

const (
	TrustSchema           = "harpy-restore-orchestrator-trust.v1"
	AcknowledgementSchema = "harpy-restore-orchestrator-ack.v1"
	Algorithm             = "Ed25519"
	MaximumDocumentBytes  = 64 * 1024
	PurposeDrain          = "drain"
	PurposeSwitch         = "switch"
)

type TrustInput struct {
	EnvironmentID           string
	DeploymentBindingSHA256 string
	Generation              int64
	KeyID                   string
	PublicKey               ed25519.PublicKey
}

type Trust struct {
	wire      trustWire
	canonical []byte
	digest    string
	publicKey ed25519.PublicKey
}

func (value Trust) EnvironmentID() string { return value.wire.EnvironmentID }
func (value Trust) DeploymentBindingSHA256() string {
	return value.wire.DeploymentBindingSHA256
}
func (value Trust) Generation() int64      { return value.wire.Generation }
func (value Trust) KeyID() string          { return value.wire.KeyID }
func (value Trust) SHA256() string         { return value.digest }
func (value Trust) CanonicalBytes() []byte { return append([]byte(nil), value.canonical...) }

type AcknowledgementInput struct {
	Purpose                       string
	EnvironmentID, OperationID    string
	AttemptID, PlanSHA256         string
	Sequence                      int64
	IssuedAt, ExpiresAt           time.Time
	PreviousDatabaseSystemID      string
	NewDatabaseSystemID           string
	PreviousDatabaseLocatorSHA256 string
	NewDatabaseLocatorSHA256      string
	StoppedServices               []string
	SwitchOutcome                 string
}

type Expected struct {
	Purpose                       string
	EnvironmentID, OperationID    string
	AttemptID, PlanSHA256         string
	PreviousDatabaseSystemID      string
	NewDatabaseSystemID           string
	PreviousDatabaseLocatorSHA256 string
	NewDatabaseLocatorSHA256      string
}

type Acknowledgement struct {
	wire      acknowledgementWire
	canonical []byte
	digest    string
	issuedAt  time.Time
	expiresAt time.Time
}

func (value Acknowledgement) Purpose() string      { return value.wire.Purpose }
func (value Acknowledgement) SHA256() string       { return value.digest }
func (value Acknowledgement) Sequence() int64      { return value.wire.Sequence }
func (value Acknowledgement) IssuedAt() time.Time  { return value.issuedAt }
func (value Acknowledgement) ExpiresAt() time.Time { return value.expiresAt }
func (value Acknowledgement) CanonicalBytes() []byte {
	return append([]byte(nil), value.canonical...)
}

type trustWire struct {
	Schema                  string `json:"schema"`
	EnvironmentID           string `json:"environment_id"`
	DeploymentBindingSHA256 string `json:"deployment_binding_sha256"`
	Generation              int64  `json:"generation"`
	KeyID                   string `json:"key_id"`
	Algorithm               string `json:"algorithm"`
	PublicKey               string `json:"public_key_base64url"`
}

type acknowledgementWire struct {
	Schema                        string   `json:"schema"`
	Purpose                       string   `json:"purpose"`
	EnvironmentID                 string   `json:"environment_id"`
	OperationID                   string   `json:"operation_id"`
	AttemptID                     string   `json:"attempt_id"`
	PlanSHA256                    string   `json:"plan_sha256"`
	Sequence                      int64    `json:"sequence"`
	IssuedAt                      string   `json:"issued_at"`
	ExpiresAt                     string   `json:"expires_at"`
	PreviousDatabaseSystemID      string   `json:"previous_database_system_id"`
	NewDatabaseSystemID           string   `json:"new_database_system_id,omitempty"`
	PreviousDatabaseLocatorSHA256 string   `json:"previous_database_locator_sha256"`
	NewDatabaseLocatorSHA256      string   `json:"new_database_locator_sha256"`
	StoppedServices               []string `json:"stopped_services,omitempty"`
	SwitchOutcome                 string   `json:"switch_outcome,omitempty"`
	KeyGeneration                 int64    `json:"key_generation"`
	KeyID                         string   `json:"key_id"`
	Algorithm                     string   `json:"algorithm"`
	Signature                     string   `json:"signature,omitempty"`
}
