package profile

import _ "embed"

//go:embed sql/tenant/publish.sql
var publishSQL string

//go:embed sql/tenant/backfill.sql
var backfillSQL string

//go:embed sql/tenant/retire.sql
var retireSQL string

//go:embed sql/tenant/get.sql
var getSelfSQL string

//go:embed sql/tenant/update.sql
var updateSelfSQL string
