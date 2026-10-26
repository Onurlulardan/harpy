package postgres

import _ "embed"

//go:embed sql/tenant/append.sql
var appendSQL string

//go:embed sql/tenant/list.sql
var listSQL string

//go:embed sql/tenant/get.sql
var getSQL string

//go:embed sql/tenant/stats.sql
var statsSQL string
