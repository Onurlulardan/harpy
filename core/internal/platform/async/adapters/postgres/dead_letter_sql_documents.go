package postgres

import _ "embed"

//go:embed sql/resolve_dead_letter_authority.sql
var resolveDeadLetterAuthoritySQL string

//go:embed sql/insert_dead_letter_event.sql
var insertDeadLetterEventSQL string

//go:embed sql/list_dead_letter_roots.sql
var listDeadLetterRootsSQL string

//go:embed sql/lock_dead_letter_root.sql
var lockDeadLetterRootSQL string

//go:embed sql/insert_dead_letter_action.sql
var insertDeadLetterActionSQL string

//go:embed sql/insert_dead_letter_event_retry.sql
var insertDeadLetterEventRetrySQL string

//go:embed sql/insert_dead_letter_job_retry.sql
var insertDeadLetterJobRetrySQL string
