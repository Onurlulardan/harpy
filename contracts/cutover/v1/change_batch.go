package cutoverv1

import "encoding/json"

const (
	MaximumChangeBatchBytes   = 8 * 1024 * 1024
	MaximumChangeBatchChanges = 1000
	MaximumChangeBatchOrdinal = 2147483
)

type SourcePosition struct {
	SnapshotName string  `json:"snapshot_name"`
	SnapshotLSN  string  `json:"snapshot_lsn"`
	ReceivedLSN  string  `json:"received_lsn"`
	CommitLSN    string  `json:"commit_lsn"`
	FinalLSN     *string `json:"final_lsn" harpyNull:"allowed"`
	SourceXID    string  `json:"source_xid"`
	BatchOrdinal int     `json:"batch_ordinal"`
}

type Change struct {
	Ordinal int             `json:"ordinal"`
	Kind    string          `json:"kind"`
	Key     json.RawMessage `json:"key"`
	Before  json.RawMessage `json:"before" harpyNull:"allowed"`
	After   json.RawMessage `json:"after" harpyNull:"allowed"`
}

func PersistedChangeOrdinal(batchOrdinal, localOrdinal int) (int, error) {
	if batchOrdinal < 1 || batchOrdinal > MaximumChangeBatchOrdinal ||
		localOrdinal < 1 || localOrdinal > MaximumChangeBatchChanges {
		return 0, contractError("persisted change ordinal")
	}
	return (batchOrdinal-1)*MaximumChangeBatchChanges + localOrdinal, nil
}

type ChangeBatchPayload struct {
	OperationID     string          `json:"operation_id"`
	RunID           string          `json:"run_id"`
	ScopeKind       string          `json:"scope_kind"`
	ScopeID         string          `json:"scope_id"`
	RelationOrdinal int             `json:"relation_ordinal"`
	CheckpointKey   string          `json:"checkpoint_key"`
	CheckpointState string          `json:"checkpoint_state"`
	Position        SourcePosition  `json:"source_position"`
	LastKey         json.RawMessage `json:"last_key" harpyNull:"allowed"`
	Changes         []Change        `json:"changes"`
}

type ChangeBatch struct {
	ContractVersion string             `json:"contract_version"`
	Payload         ChangeBatchPayload `json:"payload"`
	BatchSHA256     string             `json:"batch_sha256"`
}

type ApplyResult struct {
	AppliedCount   int64  `json:"applied_count"`
	KeyStateSHA256 string `json:"key_state_sha256"`
	StateSHA256    string `json:"state_sha256"`
}
