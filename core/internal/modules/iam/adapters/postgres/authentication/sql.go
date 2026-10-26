package postgres

import _ "embed"

//go:embed sql/tenant/login_credential.sql
var loginCredentialSQL string

//go:embed sql/tenant/use_password.sql
var usePasswordSQL string

//go:embed sql/tenant/issue_session.sql
var issueSessionSQL string

//go:embed sql/tenant/refresh_credential.sql
var refreshCredentialSQL string

//go:embed sql/tenant/rotate_refresh.sql
var rotateRefreshSQL string

//go:embed sql/platform/create_refresh_locator.sql
var createRefreshLocatorSQL string

//go:embed sql/platform/advance_refresh_locator.sql
var advanceRefreshLocatorSQL string

//go:embed sql/tenant/create_reauthentication_challenge.sql
var createReauthenticationChallengeSQL string

//go:embed sql/tenant/get_reauthentication_credential.sql
var getReauthenticationCredentialSQL string

//go:embed sql/tenant/reject_reauthentication.sql
var rejectReauthenticationSQL string

//go:embed sql/tenant/complete_reauthentication.sql
var completeReauthenticationSQL string

//go:embed sql/tenant/revoke_session.sql
var revokeSessionSQL string

//go:embed sql/tenant/resolve_recovery_credential.sql
var resolveRecoveryCredentialSQL string

//go:embed sql/tenant/consume_recovery.sql
var consumeRecoverySQL string
