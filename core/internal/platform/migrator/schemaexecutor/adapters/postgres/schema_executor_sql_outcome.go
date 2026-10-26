package postgres

import _ "embed"

//go:embed sql/insert_step_attempt.sql
var schemaExecutorInsertAttemptSQL string

//go:embed sql/finish_step.sql
var schemaExecutorFinishStepSQL string

//go:embed sql/advance_job.sql
var schemaExecutorAdvanceJobSQL string

//go:embed sql/complete_job.sql
var schemaExecutorCompleteJobSQL string

//go:embed sql/retry_job.sql
var schemaExecutorRetryJobSQL string

//go:embed sql/fail_job.sql
var schemaExecutorFailJobSQL string

//go:embed sql/fail_expired_job.sql
var schemaExecutorFailExpiredJobSQL string

//go:embed sql/select_expired_job.sql
var schemaExecutorSelectExpiredSQL string

//go:embed sql/expire_job.sql
var schemaExecutorExpireJobSQL string

//go:embed sql/recover_job.sql
var schemaExecutorRecoverJobSQL string

//go:embed sql/update_backfill_progress.sql
var schemaExecutorUpdateProgressSQL string
