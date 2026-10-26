package customerworker

import _ "embed"

//go:embed sql/platform/claim.sql
var claimSQL string
