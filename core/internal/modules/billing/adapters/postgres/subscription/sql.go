package subscription

import _ "embed"

var (
	//go:embed sql/platform/resolve_authority.sql
	resolveAuthoritySQL string
	//go:embed sql/platform/assert_authority.sql
	assertAuthoritySQL string
	//go:embed sql/platform/lock_subscription.sql
	lockSubscriptionSQL string
	//go:embed sql/platform/find_receipt.sql
	findReceiptSQL string
	//go:embed sql/platform/find_target_price.sql
	findTargetPriceSQL string
	//go:embed sql/platform/find_price_version.sql
	findPriceVersionSQL string
	//go:embed sql/platform/list_boundary_jobs.sql
	listBoundaryJobsSQL string
	//go:embed sql/platform/find_pending_change.sql
	findPendingChangeSQL string
	//go:embed sql/platform/find_trial_claim.sql
	findTrialClaimSQL string
	//go:embed sql/platform/assert_provider_fact.sql
	assertProviderFactSQL string
	//go:embed sql/platform/assert_job.sql
	assertJobSQL string
	//go:embed sql/platform/find_boundary_evidence.sql
	findBoundaryEvidenceSQL string
	//go:embed sql/platform/update_subscription.sql
	updateSubscriptionSQL string
	//go:embed sql/platform/insert_receipt.sql
	insertReceiptSQL string
	//go:embed sql/platform/insert_audit.sql
	insertAuditSQL string
	//go:embed sql/platform/insert_change.sql
	insertChangeSQL string
	//go:embed sql/platform/resolve_change.sql
	resolveChangeSQL string
	//go:embed sql/platform/insert_trial_claim.sql
	insertTrialClaimSQL string
	//go:embed sql/platform/insert_boundary_job.sql
	insertBoundaryJobSQL string
	//go:embed sql/platform/complete_provider_receipt.sql
	completeProviderReceiptSQL string
)
