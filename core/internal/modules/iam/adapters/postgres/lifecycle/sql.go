package postgres

import _ "embed"

//go:embed sql/tenant/credential_lookup.sql
var credentialLookupSQL string

//go:embed sql/tenant/email_delivery_target.sql
var emailDeliveryTargetSQL string

//go:embed sql/tenant/password_delivery_target.sql
var passwordDeliveryTargetSQL string

//go:embed sql/tenant/portal_delivery_target.sql
var portalDeliveryTargetSQL string

//go:embed sql/tenant/identifier_add.sql
var identifierAddSQL string

//go:embed sql/tenant/identifier_verify.sql
var identifierVerifySQL string

//go:embed sql/tenant/identifier_tombstone.sql
var identifierTombstoneSQL string

//go:embed sql/tenant/identifier_retire.sql
var identifierRetireSQL string

//go:embed sql/tenant/email_verification_start.sql
var emailVerificationStartSQL string

//go:embed sql/tenant/password_reset_start.sql
var passwordResetStartSQL string

//go:embed sql/tenant/password_reset_complete.sql
var passwordResetCompleteSQL string

//go:embed sql/tenant/initial_admin_setup_accept.sql
var initialAdminSetupAcceptSQL string

//go:embed sql/tenant/portal_setup_start.sql
var portalSetupStartSQL string

//go:embed sql/tenant/portal_setup_complete.sql
var portalSetupCompleteSQL string

//go:embed sql/tenant/portal_setup_read.sql
var portalSetupReadSQL string

//go:embed sql/tenant/portal_approval_decide.sql
var portalApprovalDecideSQL string

//go:embed sql/tenant/password_credential_read.sql
var passwordCredentialReadSQL string

//go:embed sql/tenant/password_change.sql
var passwordChangeSQL string

//go:embed sql/tenant/authenticator_list.sql
var authenticatorListSQL string

//go:embed sql/tenant/authenticator_suspend.sql
var authenticatorSuspendSQL string

//go:embed sql/tenant/authenticator_retire.sql
var authenticatorRetireSQL string

//go:embed sql/tenant/webauthn_start.sql
var webAuthnStartSQL string

//go:embed sql/tenant/webauthn_complete.sql
var webAuthnCompleteSQL string

//go:embed sql/tenant/totp_start.sql
var totpStartSQL string

//go:embed sql/tenant/totp_read.sql
var totpReadSQL string

//go:embed sql/tenant/totp_complete.sql
var totpCompleteSQL string

//go:embed sql/tenant/recovery_codes_issue.sql
var recoveryCodesIssueSQL string

//go:embed sql/tenant/recovery_codes_list.sql
var recoveryCodesListSQL string

//go:embed sql/tenant/session_list.sql
var sessionListSQL string

//go:embed sql/tenant/session_revoke_one.sql
var sessionRevokeOneSQL string

//go:embed sql/tenant/session_revoke_all.sql
var sessionRevokeAllSQL string
