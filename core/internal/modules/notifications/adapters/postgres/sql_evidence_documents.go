package postgres

import _ "embed"

var (
	//go:embed sql/tenant/append_audit.sql
	appendAuditSQL string
	//go:embed sql/tenant/append_realtime_hint.sql
	appendRealtimeHintSQL string
	//go:embed sql/tenant/append_operator_alert.sql
	appendOperatorAlertSQL string
	//go:embed sql/tenant/claim_realtime_hints.sql
	claimRealtimeHintsSQL string
)
