package jobs

import _ "embed"

//go:embed sql/tenant/load_recurrence.sql
var loadRecurrenceSQL string

//go:embed sql/tenant/count_future_occurrences.sql
var countFutureSQL string

//go:embed sql/tenant/insert_occurrence.sql
var insertOccurrenceSQL string
