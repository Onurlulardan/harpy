package application

const (
	PermissionRead Permission = "changelog.read"

	OperationList              = "changelog.list"
	OperationStats             = "changelog.stats"
	OperationGet               = "changelog.get"
	OperationEntityHistory     = "changelog.entity_history"
	OperationFormRecordHistory = "changelog.form_record_history"
)

type Permission string

type ScopeAccess uint8

const (
	ScopeCommand ScopeAccess = iota + 1
	ScopeSnapshot
)
