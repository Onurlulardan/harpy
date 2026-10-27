package postgres

import _ "embed"

//go:embed sql/tenant/initial_password_credential.sql
var initialPasswordCredentialSQL string

//go:embed sql/tenant/initial_password_start.sql
var initialPasswordStartSQL string

//go:embed sql/tenant/initial_password_continuation.sql
var initialPasswordContinuationSQL string

//go:embed sql/tenant/initial_password_reject.sql
var initialPasswordRejectSQL string

//go:embed sql/tenant/initial_password_complete.sql
var initialPasswordCompleteSQL string
