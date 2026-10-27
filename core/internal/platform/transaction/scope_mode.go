package transaction

type ScopeMode string

const (
	Command       ScopeMode = "command"
	SnapshotQuery ScopeMode = "snapshot_query"
)

func (mode ScopeMode) valid() bool {
	return mode == Command || mode == SnapshotQuery
}
