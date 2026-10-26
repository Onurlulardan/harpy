package postgres

import _ "embed"

//go:embed sql/run_retention_cycle.sql
var runRetentionCycleSQL string
