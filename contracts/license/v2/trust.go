package licensev2

import "encoding/json"

type TrustBundle struct {
	Schema  string           `json:"schema"`
	Purpose string           `json:"purpose"`
	Roots   []TrustRoot      `json:"roots"`
	Keysets []KeysetEnvelope `json:"keysets"`
}

type TrustRoot struct {
	KeyID       string         `json:"kid"`
	Algorithm   string         `json:"algorithm"`
	PublicKey   string         `json:"public_key"`
	Status      TrustKeyStatus `json:"status"`
	NotBefore   string         `json:"not_before"`
	SignUntil   string         `json:"sign_until"`
	VerifyUntil string         `json:"verify_until"`
}

type KeysetEnvelope struct {
	Schema           string          `json:"schema"`
	KeyID            string          `json:"kid"`
	Algorithm        string          `json:"algorithm"`
	Canonicalization string          `json:"canonicalization"`
	PayloadSHA256    string          `json:"payload_sha256"`
	Payload          json.RawMessage `json:"payload"`
	Signature        string          `json:"signature"`
}

type KeysetPayload struct {
	Schema          string           `json:"schema"`
	Issuer          string           `json:"issuer"`
	Audience        string           `json:"audience"`
	Sequence        int64            `json:"sequence"`
	IssuedAt        string           `json:"issued_at"`
	NotBefore       string           `json:"not_before"`
	ExpiresAt       string           `json:"expires_at"`
	SnapshotSigners []SnapshotSigner `json:"snapshot_signers"`
}

type SnapshotSigner struct {
	KeyID       string         `json:"kid"`
	Algorithm   string         `json:"algorithm"`
	PublicKey   string         `json:"public_key"`
	Status      TrustKeyStatus `json:"status"`
	NotBefore   string         `json:"not_before"`
	SignUntil   string         `json:"sign_until"`
	VerifyUntil string         `json:"verify_until"`
}
