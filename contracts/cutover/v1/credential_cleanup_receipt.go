package cutoverv1

type CleanedPrincipal struct {
	Principal            string  `json:"principal"`
	CleanedAt            string  `json:"cleaned_at"`
	CanLogin             bool    `json:"can_login"`
	PasswordState        *string `json:"password_state" harpyNull:"allowed"`
	Replication          bool    `json:"replication"`
	ConnectionLimit      int     `json:"connection_limit"`
	ActiveSessions       int     `json:"active_sessions"`
	MembershipCount      int     `json:"membership_count"`
	DynamicGrantsRevoked bool    `json:"dynamic_grants_revoked"`
}

type CredentialCleanupPayload struct {
	ReceiptID           string             `json:"receipt_id"`
	LeaseID             string             `json:"lease_id"`
	LeaseManifestSHA256 string             `json:"lease_manifest_sha256"`
	SealReceiptSHA256   string             `json:"seal_receipt_sha256"`
	EnvironmentID       string             `json:"environment_id"`
	OperationID         string             `json:"operation_id"`
	RunID               string             `json:"run_id"`
	IssuedAt            string             `json:"issued_at"`
	Principals          []CleanedPrincipal `json:"principals"`
	Result              string             `json:"result"`
}

type CredentialCleanupReceipt struct {
	ContractVersion string                   `json:"contract_version"`
	Kind            string                   `json:"kind"`
	Payload         CredentialCleanupPayload `json:"payload"`
	Signature       Signature                `json:"signature"`
}

type CredentialCleanupSigningEnvelope struct {
	ContractVersion string                   `json:"contract_version"`
	Kind            string                   `json:"kind"`
	Payload         CredentialCleanupPayload `json:"payload"`
}

func (receipt CredentialCleanupReceipt) SigningEnvelope() CredentialCleanupSigningEnvelope {
	return CredentialCleanupSigningEnvelope{
		ContractVersion: receipt.ContractVersion, Kind: receipt.Kind, Payload: receipt.Payload,
	}
}
