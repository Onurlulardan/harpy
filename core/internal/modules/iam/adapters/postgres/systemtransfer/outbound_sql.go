package systemtransfer

import _ "embed"

var (
	//go:embed sql/tenant/load_export_worker.sql
	outboundLoadSQL string
	//go:embed sql/tenant/claim_export_worker.sql
	outboundClaimSQL string
	//go:embed sql/tenant/checkpoint_export_worker.sql
	outboundCheckpointSQL string
	//go:embed sql/tenant/start_export_effect.sql
	outboundEffectStartSQL string
	//go:embed sql/tenant/complete_export_worker.sql
	outboundCompleteSQL string
	//go:embed sql/tenant/finish_export_worker.sql
	outboundFinishSQL string
	//go:embed sql/tenant/mark_export_unknown.sql
	outboundUnknownSQL string
)
