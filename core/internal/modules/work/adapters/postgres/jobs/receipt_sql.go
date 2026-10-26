package jobs

import _ "embed"

//go:embed sql/tenant/begin_job.sql
var beginJobSQL string

//go:embed sql/tenant/load_job.sql
var loadJobSQL string

//go:embed sql/tenant/complete_job.sql
var completeJobSQL string
