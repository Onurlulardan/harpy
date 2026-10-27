package cutoverv1

const MaximumTerminalEvidenceBytes = 16 * 1024 * 1024

type TerminalEvidence struct {
	ContractVersion           string                       `json:"contract_version"`
	OperationID               string                       `json:"operation_id"`
	RunID                     string                       `json:"run_id"`
	LeaseManifestSHA256       string                       `json:"lease_manifest_sha256"`
	PlanSHA256                string                       `json:"plan_sha256"`
	TransformManifestSHA256   string                       `json:"transform_manifest_sha256"`
	TerminalState             string                       `json:"terminal_state"`
	FinalLSN                  *string                      `json:"final_lsn,omitempty" harpyOptional:"allowed"`
	InitialCopy               *TerminalInitialCopyEvidence `json:"initial_copy,omitempty" harpyOptional:"allowed"`
	CheckpointCount           *int                         `json:"checkpoint_count,omitempty" harpyOptional:"allowed"`
	Checkpoints               []TerminalCheckpointEvidence `json:"checkpoints,omitempty" harpyOptional:"allowed"`
	JournalTransactionCount   *int64                       `json:"journal_transaction_count,omitempty" harpyOptional:"allowed"`
	AppliedChangeCount        *int64                       `json:"applied_change_count,omitempty" harpyOptional:"allowed"`
	JournalCoverageRootSHA256 *string                      `json:"journal_coverage_root_sha256,omitempty" harpyOptional:"allowed"`
	OutcomeCode               *string                      `json:"outcome_code,omitempty" harpyOptional:"allowed"`
	OutcomeEvidenceSHA256     *string                      `json:"outcome_evidence_sha256,omitempty" harpyOptional:"allowed"`
}

type TerminalInitialCopyEvidence struct {
	ManifestSHA256 string `json:"manifest_sha256"`
	SnapshotName   string `json:"snapshot_name"`
	SnapshotLSN    string `json:"snapshot_lsn"`
	RelationCount  int    `json:"relation_count"`
	PartitionCount int    `json:"partition_count"`
	RowCount       int64  `json:"row_count"`
	RootSHA256     string `json:"root_sha256"`
}

type TerminalCheckpointEvidence struct {
	RelationOrdinal int    `json:"relation_ordinal"`
	CheckpointKey   string `json:"checkpoint_key"`
	AppliedLSN      string `json:"applied_lsn"`
	AppliedRows     int64  `json:"applied_rows"`
	BatchCount      int    `json:"batch_count"`
	BatchRootSHA256 string `json:"batch_root_sha256"`
}

type TerminalBatchLeaf struct {
	BatchOrdinal    int
	SourceXID       string
	SourceCommitLSN string
	BatchSHA256     string
	AppliedCount    int64
	KeyStateSHA256  string
	StateSHA256     string
}

type TerminalChangeLeaf struct {
	RelationOrdinal int
	ChangeOrdinal   int
	CheckpointKey   string
	MutationKind    string
	MutationSHA256  string
}

type TerminalJournalCoverageLeaf struct {
	SourceXID       string
	SourceCommitLSN string
	Changes         []TerminalChangeLeaf
}
