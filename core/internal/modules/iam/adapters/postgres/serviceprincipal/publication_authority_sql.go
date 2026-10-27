package serviceprincipal

import _ "embed"

//go:embed sql/tenant/publication_authority_permission.sql
var publicationAuthorityPermissionSQL string

//go:embed sql/tenant/publication_authority_principal.sql
var publicationAuthorityPrincipalSQL string

//go:embed sql/tenant/publication_authority_grant.sql
var publicationAuthorityGrantSQL string
