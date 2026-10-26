package postgres

import _ "embed"

//go:embed sql/tenant/permission_request_create.sql
var permissionRequestCreateSQL string

//go:embed sql/tenant/permission_request_authority.sql
var permissionRequestAuthoritySQL string

//go:embed sql/tenant/permission_request_read.sql
var permissionRequestReadSQL string

//go:embed sql/tenant/permission_request_list.sql
var permissionRequestListSQL string

//go:embed sql/tenant/permission_request_review.sql
var permissionRequestReviewSQL string

//go:embed sql/tenant/permission_request_cancel.sql
var permissionRequestCancelSQL string

//go:embed sql/tenant/permission_request_expire.sql
var permissionRequestExpireSQL string

//go:embed sql/tenant/permission_request_decide.sql
var permissionRequestDecideSQL string

//go:embed sql/tenant/permission_request_decision_audit.sql
var permissionRequestDecisionAuditSQL string

//go:embed sql/tenant/permission_request_transition_audit.sql
var permissionRequestTransitionAuditSQL string

//go:embed sql/tenant/permission_request_catalog_resources.sql
var permissionRequestCatalogResourcesSQL string

//go:embed sql/tenant/permission_request_catalog_actions.sql
var permissionRequestCatalogActionsSQL string

//go:embed sql/tenant/permission_request_catalog_missing_resources.sql
var permissionRequestCatalogMissingResourcesSQL string
