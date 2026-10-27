package postgres

import _ "embed"

//go:embed sql/select_due_job.sql
var schemaExecutorSelectDueJobSQL string

//go:embed sql/lease_job.sql
var schemaExecutorLeaseJobSQL string

//go:embed sql/start_step.sql
var schemaExecutorStartStepSQL string

//go:embed sql/read_job_steps.sql
var schemaExecutorReadJobStepsSQL string

//go:embed sql/mark_job_running.sql
var schemaExecutorMarkRunningSQL string

//go:embed sql/heartbeat_job.sql
var schemaExecutorHeartbeatSQL string

//go:embed sql/lock_active_lease.sql
var schemaExecutorLockLeaseSQL string

//go:embed sql/check_active_lease.sql
var schemaExecutorCheckLeaseSQL string
