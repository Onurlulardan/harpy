package catalogowner

import _ "embed"

//go:embed sql/tenant/provision.sql
var provisionSQL string
