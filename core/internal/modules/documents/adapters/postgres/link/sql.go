package link

import _ "embed"

//go:embed sql/tenant/register.sql
var registerSQL string

//go:embed sql/tenant/release.sql
var releaseSQL string

//go:embed sql/tenant/list.sql
var listSQL string

//go:embed sql/tenant/lock_active_epoch.sql
var lockActiveEpochSQL string
