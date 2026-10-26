package cutoverv1

type SealedCredential struct {
	Principal       string  `json:"principal"`
	CanLogin        bool    `json:"can_login"`
	PasswordState   *string `json:"password_state" harpyNull:"allowed"`
	Replication     bool    `json:"replication"`
	ConnectionLimit int     `json:"connection_limit"`
}

type CutoverSession struct {
	PID             int64  `json:"pid"`
	BackendStart    string `json:"backend_start"`
	SessionUser     string `json:"session_user"`
	ApplicationName string `json:"application_name"`
}

type SessionSealPayload struct {
	ReceiptID           string             `json:"receipt_id"`
	LeaseID             string             `json:"lease_id"`
	LeaseManifestSHA256 string             `json:"lease_manifest_sha256"`
	EnvironmentID       string             `json:"environment_id"`
	OperationID         string             `json:"operation_id"`
	RunID               string             `json:"run_id"`
	IssuedAt            string             `json:"issued_at"`
	Source              DatabaseIdentity   `json:"source"`
	Target              DatabaseIdentity   `json:"target"`
	Credentials         []SealedCredential `json:"credentials"`
	SourceSessions      []CutoverSession   `json:"source_sessions"`
	TargetSessions      []CutoverSession   `json:"target_sessions"`
	Result              string             `json:"result"`
}

type SessionSealReceipt struct {
	ContractVersion string             `json:"contract_version"`
	Kind            string             `json:"kind"`
	Payload         SessionSealPayload `json:"payload"`
	Signature       Signature          `json:"signature"`
}

type SessionSealSigningEnvelope struct {
	ContractVersion string             `json:"contract_version"`
	Kind            string             `json:"kind"`
	Payload         SessionSealPayload `json:"payload"`
}

func (receipt SessionSealReceipt) SigningEnvelope() SessionSealSigningEnvelope {
	return SessionSealSigningEnvelope{
		ContractVersion: receipt.ContractVersion, Kind: receipt.Kind, Payload: receipt.Payload,
	}
}
