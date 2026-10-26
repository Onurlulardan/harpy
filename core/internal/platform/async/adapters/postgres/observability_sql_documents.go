package postgres

import _ "embed"

//go:embed sql/observe_async_backlog.sql
var observeAsyncBacklogSQL string
