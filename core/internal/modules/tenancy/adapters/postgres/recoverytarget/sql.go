package recoverytarget

import _ "embed"

//go:embed sql/platform/get.sql
var getRecoveryTargetSQL string
