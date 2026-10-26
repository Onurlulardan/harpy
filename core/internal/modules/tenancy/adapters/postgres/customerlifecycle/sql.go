package customerlifecycle

import _ "embed"

//go:embed sql/platform/list.sql
var listSQL string

//go:embed sql/platform/get.sql
var getSQL string

//go:embed sql/platform/mutate.sql
var mutateSQL string

//go:embed sql/platform/accept.sql
var acceptSQL string

//go:embed sql/platform/read_operation.sql
var readOperationSQL string

//go:embed sql/platform/claim.sql
var claimSQL string

//go:embed sql/platform/complete_offboard.sql
var completeOffboardSQL string

//go:embed sql/platform/complete_purge.sql
var completePurgeSQL string

//go:embed sql/platform/fail.sql
var failSQL string
