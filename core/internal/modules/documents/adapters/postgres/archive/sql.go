package archive

import _ "embed"

//go:embed sql/tenant/create.sql
var createSQL string

//go:embed sql/tenant/find_create.sql
var findCreateSQL string

//go:embed sql/tenant/get.sql
var getSQL string

//go:embed sql/tenant/pin.sql
var pinSQL string

//go:embed sql/tenant/entries.sql
var entriesSQL string

//go:embed sql/tenant/record_effect.sql
var recordEffectSQL string

//go:embed sql/tenant/mark_ready.sql
var markReadySQL string

//go:embed sql/tenant/expire.sql
var expireSQL string
