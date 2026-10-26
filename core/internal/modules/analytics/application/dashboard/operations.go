package dashboard

const (
	OperationCreate     = "analytics.dashboard.create"
	OperationGet        = "analytics.dashboard.get"
	OperationList       = "analytics.dashboard.list"
	OperationUpdate     = "analytics.dashboard.update"
	OperationDelete     = "analytics.dashboard.delete"
	OperationPublish    = "analytics.dashboard.publish"
	OperationRead       = "analytics.dashboard.read"
	OperationGrantACL   = "analytics.dashboard.grant_acl"
	OperationRevokeACL  = "analytics.dashboard.revoke_acl"
	OperationSetDefault = "analytics.dashboard.set_default"
)

type Permission string

const (
	PermissionCreate  Permission = "analytics.dashboard.create"
	PermissionRead    Permission = "analytics.dashboard.read"
	PermissionUpdate  Permission = "analytics.dashboard.update"
	PermissionDelete  Permission = "analytics.dashboard.delete"
	PermissionPublish Permission = "analytics.dashboard.publish"
)
