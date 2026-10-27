package aicredentialv1

import "errors"

const (
	ContractVersion       = "ai-credential-envelope.v1"
	Audience              = "harpy-ai-provider-v1"
	MaximumEnvelopeBytes  = 16 << 10
	MaximumPlaintextBytes = 8 << 10
)

var ErrInvalidEnvelope = errors.New("AI credential envelope is invalid")

// AAD is the complete authenticated context for one sealed provider credential.
type AAD struct {
	TenantID       string `json:"tenant_id"`
	ConfigID       string `json:"config_id"`
	ConfigRevision int64  `json:"config_revision"`
	Provider       string `json:"provider"`
	Audience       string `json:"audience"`
	KeyID          string `json:"key_id"`
}

type Envelope struct {
	ContractVersion    string `json:"contract_version"`
	AAD                AAD    `json:"aad"`
	EphemeralPublicKey string `json:"ephemeral_public_key"`
	Nonce              string `json:"nonce"`
	Ciphertext         string `json:"ciphertext"`
}
