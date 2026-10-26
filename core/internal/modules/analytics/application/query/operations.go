package query

const (
	OperationCreate       = "analytics.query.create"
	OperationGet          = "analytics.query.get"
	OperationGetPublished = OperationGet
	OperationList         = "analytics.query.list"
	OperationUpdate       = "analytics.query.update"
	OperationDelete       = "analytics.query.delete"
	OperationPublish      = "analytics.query.publish"
	OperationExecute      = "analytics.query.execute"
	OperationRevalidate   = "analytics.query.revalidate"
)

type Permission string

const (
	PermissionCreate  Permission = "analytics.query.create"
	PermissionRead    Permission = "analytics.query.read"
	PermissionUpdate  Permission = "analytics.query.update"
	PermissionDelete  Permission = "analytics.query.delete"
	PermissionPublish Permission = "analytics.query.publish"
	PermissionExecute Permission = "analytics.query.execute"
)
