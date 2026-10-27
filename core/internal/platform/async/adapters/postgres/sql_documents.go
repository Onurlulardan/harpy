package postgres

import _ "embed"

//go:embed sql/enqueue_job.sql
var enqueueJobSQL string

//go:embed sql/find_job_enqueue.sql
var findJobEnqueueSQL string

//go:embed sql/enqueue_event.sql
var enqueueEventSQL string

//go:embed sql/find_event_enqueue.sql
var findEventEnqueueSQL string

//go:embed sql/lock_job_claim.sql
var lockJobClaimSQL string

//go:embed sql/heartbeat_job.sql
var heartbeatJobSQL string

//go:embed sql/mark_external_job_effect.sql
var markExternalJobEffectSQL string

//go:embed sql/mark_external_attempt_effect.sql
var markExternalAttemptEffectSQL string

//go:embed sql/observe_retried_external_effect.sql
var observeRetriedExternalEffectSQL string

//go:embed sql/complete_job_attempt.sql
var completeJobAttemptSQL string

//go:embed sql/complete_database_job.sql
var completeDatabaseJobSQL string

//go:embed sql/store_external_effect_receipt.sql
var storeExternalEffectReceiptSQL string

//go:embed sql/complete_external_job_attempt.sql
var completeExternalJobAttemptSQL string

//go:embed sql/complete_external_job.sql
var completeExternalJobSQL string

//go:embed sql/has_external_effect_receipt.sql
var hasExternalEffectReceiptSQL string

//go:embed sql/mark_external_attempt_unknown.sql
var markExternalAttemptUnknownSQL string

//go:embed sql/mark_external_job_unknown.sql
var markExternalJobUnknownSQL string

//go:embed sql/fail_job_attempt.sql
var failJobAttemptSQL string

//go:embed sql/fail_job.sql
var failJobSQL string

//go:embed sql/resolve_job_control.sql
var resolveJobControlSQL string

//go:embed sql/check_job_active.sql
var checkJobActiveSQL string

//go:embed sql/request_job_cancellation.sql
var requestJobCancellationSQL string

//go:embed sql/cancel_owned_job.sql
var cancelOwnedJobSQL string

//go:embed sql/lock_expired_jobs.sql
var lockExpiredJobsSQL string

//go:embed sql/recover_expired_job_attempt.sql
var recoverExpiredJobAttemptSQL string

//go:embed sql/recover_expired_job.sql
var recoverExpiredJobSQL string

//go:embed sql/park_stale_job.sql
var parkStaleJobSQL string

//go:embed sql/list_stale_jobs.sql
var listStaleJobsSQL string

//go:embed sql/list_stale_external_jobs.sql
var listStaleExternalJobsSQL string

//go:embed sql/finalize_stale_external_job.sql
var finalizeStaleExternalJobSQL string

//go:embed sql/list_job_authorities.sql
var listJobAuthoritiesSQL string

//go:embed sql/assert_job_authority.sql
var assertJobAuthoritySQL string

//go:embed sql/claim_jobs.sql
var claimJobsSQL string

//go:embed sql/reconcile_event_gap.sql
var reconcileEventGapSQL string

//go:embed sql/resolve_stale_job_authority.sql
var resolveStaleJobAuthoritySQL string

//go:embed sql/inspect_stale_job_disposition.sql
var inspectStaleJobDispositionSQL string

//go:embed sql/dispose_stale_job.sql
var disposeStaleJobSQL string

//go:embed sql/assert_live_job_catalog.sql
var assertLiveJobCatalogSQL string

//go:embed sql/assert_live_outbox_catalog.sql
var assertLiveOutboxCatalogSQL string

//go:embed sql/insert_job_attempts.sql
var insertJobAttemptsSQL string

//go:embed sql/release_job_retention.sql
var releaseJobRetentionSQL string

//go:embed sql/read_http_idempotency_receipt.sql
var readHTTPIdempotencyReceiptSQL string

//go:embed sql/insert_http_idempotency_receipt.sql
var insertHTTPIdempotencyReceiptSQL string
