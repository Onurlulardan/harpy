package devjournal

type journalRecord struct {
	Kind               string  `json:"kind"`
	RegistrySequence   int64   `json:"registry_sequence"`
	EnvironmentID      *string `json:"environment_id"`
	RegistryBindingID  string  `json:"registry_binding_id"`
	ScopeKind          *string `json:"scope_kind"`
	ScopeID            *string `json:"scope_id"`
	OperationID        *string `json:"operation_id"`
	ProposalSHA256     *string `json:"proposal_sha256"`
	Generation         *int64  `json:"generation"`
	PreviousHeadSHA256 *string `json:"previous_head_sha256"`
	StateSHA256        *string `json:"state_sha256"`
	ProposalArtifact   *string `json:"proposal_artifact"`
	Artifact           string  `json:"artifact"`
}
