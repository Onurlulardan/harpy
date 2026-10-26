package formscleanup

import _ "embed"

//go:embed sql/tenant/create.sql
var createSQL string

//go:embed sql/tenant/prepare.sql
var prepareSQL string

//go:embed sql/tenant/stage_enqueue.sql
var stageEnqueueSQL string

//go:embed sql/tenant/mark_enqueued.sql
var markEnqueuedSQL string

//go:embed sql/tenant/finalize.sql
var finalizeSQL string
