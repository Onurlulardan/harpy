package catalogowner

import _ "embed"

//go:embed sql/tenant/provision.sql
var provisionSQL string

//go:embed sql/tenant/demo_handoff.sql
var demoHandoffSQL string
