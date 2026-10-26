package postgres

import _ "embed"

//go:embed sql/record_full_rollout_completion.sql
var recordFullRolloutCompletionSQL string
