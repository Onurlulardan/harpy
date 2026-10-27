package onboarding

import _ "embed"

//go:embed sql/platform/accept.sql
var acceptSQL string

//go:embed sql/platform/claim.sql
var claimSQL string

//go:embed sql/platform/record.sql
var recordSQL string

//go:embed sql/platform/record_authority_ack.sql
var recordAuthorityAckSQL string

//go:embed sql/platform/activate.sql
var activateSQL string

//go:embed sql/platform/complete.sql
var completeSQL string

//go:embed sql/platform/reconcile.sql
var reconcileSQL string
