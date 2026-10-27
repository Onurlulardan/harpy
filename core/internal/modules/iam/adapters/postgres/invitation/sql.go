package postgres

import _ "embed"

//go:embed sql/tenant/invitation_create.sql
var invitationCreateSQL string

//go:embed sql/tenant/invitation_read.sql
var invitationReadSQL string

//go:embed sql/tenant/invitation_list.sql
var invitationListSQL string

//go:embed sql/tenant/invitation_revoke.sql
var invitationRevokeSQL string

//go:embed sql/tenant/invitation_expire.sql
var invitationExpireSQL string

//go:embed sql/tenant/invitation_accept.sql
var invitationAcceptSQL string

//go:embed sql/tenant/invitation_role_lookup.sql
var invitationRoleLookupSQL string

//go:embed sql/tenant/invitation_credential_lookup.sql
var invitationCredentialLookupSQL string

//go:embed sql/platform/invitation_directory_upsert.sql
var invitationDirectoryUpsertSQL string
