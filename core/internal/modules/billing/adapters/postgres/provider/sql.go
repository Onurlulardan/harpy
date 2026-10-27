package provider

import _ "embed"

var (
	//go:embed sql/platform/resolve_provider_target.sql
	resolveProviderTargetSQL string
	//go:embed sql/platform/find_webhook_receipt.sql
	findWebhookReceiptSQL string
	//go:embed sql/platform/insert_webhook_receipt.sql
	insertWebhookReceiptSQL string
	//go:embed sql/platform/resolve_tenant_authority.sql
	resolveTenantAuthoritySQL string
	//go:embed sql/platform/load_webhook.sql
	loadWebhookSQL string
	//go:embed sql/platform/mark_webhook_processing.sql
	markWebhookProcessingSQL string
	//go:embed sql/platform/find_provider_event.sql
	findProviderEventSQL string
	//go:embed sql/platform/find_latest_provider_sequence.sql
	findLatestProviderSequenceSQL string
	//go:embed sql/platform/insert_provider_event.sql
	insertProviderEventSQL string
	//go:embed sql/platform/mark_webhook_reconcile.sql
	markWebhookReconcileSQL string
	//go:embed sql/platform/insert_boundary_evidence.sql
	insertBoundaryEvidenceSQL string
	//go:embed sql/platform/resolve_command_authority.sql
	resolveCommandAuthoritySQL string
	//go:embed sql/platform/create_command.sql
	createCommandSQL string
	//go:embed sql/platform/finalize_command.sql
	finalizeCommandSQL string
)
