package resource

import _ "embed"

//go:embed sql/tenant/manifest.sql
var manifestSQL string
