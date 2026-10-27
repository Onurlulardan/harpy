package publication

const (
	OperationCreate   = "analytics.publication.create"
	OperationSubmit   = "analytics.publication.submit"
	OperationApprove  = "analytics.publication.approve"
	OperationActivate = "analytics.publication.activate"
	OperationGet      = "analytics.publication.get"
	OperationList     = "analytics.publication.list"
	OperationRevoke   = "analytics.publication.revoke"
	OperationExecute  = "analytics.publication.execute"
)

type Permission string

const (
	PermissionRead              Permission = "analytics.publication.read"
	PermissionRevoke            Permission = "analytics.publication.revoke"
	PermissionPublishPublic     Permission = "analytics.publication.publish_public"
	PermissionApprovePublic     Permission = "analytics.publication.approve_public"
	PermissionPublishCapability Permission = "analytics.publication.publish_capability"
	PermissionApproveCapability Permission = "analytics.publication.approve_capability"
)
