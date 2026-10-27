package purge

import _ "embed"

//go:embed sql/tenant/lock.sql
var lockSQL string

//go:embed sql/tenant/begin.sql
var beginSQL string

//go:embed sql/tenant/finish.sql
var finishSQL string

//go:embed sql/tenant/list_pending.sql
var listPendingSQL string
