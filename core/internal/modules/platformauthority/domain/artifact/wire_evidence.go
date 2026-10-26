package artifact

type BindingProof struct {
	Schema                    string `json:"schema"`
	EnvironmentID             string `json:"environment_id"`
	LicenseBindingID          string `json:"license_binding_id"`
	FinalizationReceiptSHA256 string `json:"finalization_receipt_sha256"`
	RecoveryTrustBundleSHA256 string `json:"recovery_trust_bundle_sha256"`
	ControlKID                string `json:"control_kid"`
	RegistryNonce             string `json:"registry_nonce"`
	OperationID               string `json:"operation_id"`
	Signature                 string `json:"signature"`
}

type Checkpoint struct {
	Schema           string       `json:"schema"`
	RegistryKind     RegistryKind `json:"registry_kind"`
	RegistrySequence int64        `json:"registry_sequence"`
	ActiveScopeCount int64        `json:"active_scope_count"`
	MerkleRootSHA256 string       `json:"merkle_root_sha256"`
	IssuedAt         string       `json:"issued_at"`
	RegistryKID      string       `json:"registry_kid"`
	Signature        string       `json:"signature"`
}

type TrustKey struct {
	KID       string `json:"kid"`
	Algorithm string `json:"algorithm"`
	PublicKey string `json:"public_key"`
	NotBefore string `json:"not_before"`
	SignUntil string `json:"sign_until"`
}

type RetiredTrustKey struct {
	KID         string `json:"kid"`
	Algorithm   string `json:"algorithm"`
	PublicKey   string `json:"public_key"`
	SignUntil   string `json:"sign_until"`
	VerifyUntil string `json:"verify_until"`
	Disposition string `json:"disposition"`
}

type TrustBundle struct {
	Schema           string            `json:"schema"`
	RegistryKind     RegistryKind      `json:"registry_kind"`
	TrustGeneration  int64             `json:"trust_generation"`
	Active           TrustKey          `json:"active"`
	Retired          []RetiredTrustKey `json:"retired"`
	BundleSHA256     string            `json:"bundle_sha256"`
	ReleaseSignature string            `json:"release_signature"`
}
