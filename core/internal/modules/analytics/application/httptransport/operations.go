package httptransport

const (
	OperationQueryCreate            = "analytics.query.create"
	OperationQueryGet               = "analytics.query.get"
	OperationQueryGetPublished      = OperationQueryGet
	OperationQueryList              = "analytics.query.list"
	OperationQueryUpdate            = "analytics.query.update"
	OperationQueryDelete            = "analytics.query.delete"
	OperationQueryPublish           = "analytics.query.publish"
	OperationQueryRevalidate        = "analytics.query.revalidate"
	OperationQueryExecute           = "analytics.query.execute"
	OperationDashboardCreate        = "analytics.dashboard.create"
	OperationDashboardGet           = "analytics.dashboard.get"
	OperationDashboardList          = "analytics.dashboard.list"
	OperationDashboardUpdate        = "analytics.dashboard.update"
	OperationDashboardDelete        = "analytics.dashboard.delete"
	OperationDashboardPublish       = "analytics.dashboard.publish"
	OperationDashboardRead          = "analytics.dashboard.read"
	OperationDashboardGrantACL      = "analytics.dashboard.grant_acl"
	OperationDashboardRevokeACL     = "analytics.dashboard.revoke_acl"
	OperationDashboardSetDefault    = "analytics.dashboard.set_default"
	OperationDashboardExportCreate  = "analytics.dashboard.export.create"
	OperationDashboardExportGet     = "analytics.dashboard.export.get"
	OperationDashboardExportDelete  = "analytics.dashboard.export.delete"
	OperationDashboardExportContent = "analytics.dashboard.export.content"
	OperationPublicationCreate      = "analytics.publication.create"
	OperationPublicationGet         = "analytics.publication.get"
	OperationPublicationList        = "analytics.publication.list"
	OperationPublicationSubmit      = "analytics.publication.submit"
	OperationPublicationApprove     = "analytics.publication.approve"
	OperationPublicationActivate    = "analytics.publication.activate"
	OperationPublicationRevoke      = "analytics.publication.revoke"
)

const (
	PermissionQueryCreate                  Permission = "analytics.query.create"
	PermissionQueryRead                    Permission = "analytics.query.read"
	PermissionQueryUpdate                  Permission = "analytics.query.update"
	PermissionQueryDelete                  Permission = "analytics.query.delete"
	PermissionQueryPublish                 Permission = "analytics.query.publish"
	PermissionQueryExecute                 Permission = "analytics.query.execute"
	PermissionDashboardCreate              Permission = "analytics.dashboard.create"
	PermissionDashboardRead                Permission = "analytics.dashboard.read"
	PermissionDashboardUpdate              Permission = "analytics.dashboard.update"
	PermissionDashboardDelete              Permission = "analytics.dashboard.delete"
	PermissionDashboardPublish             Permission = "analytics.dashboard.publish"
	PermissionDashboardExport              Permission = "analytics.dashboard.export"
	PermissionPublicationRead              Permission = "analytics.publication.read"
	PermissionPublicationRevoke            Permission = "analytics.publication.revoke"
	PermissionPublicationPublishPublic     Permission = "analytics.publication.publish_public"
	PermissionPublicationPublishCapability Permission = "analytics.publication.publish_capability"
)

func validPermissionOperation(permission Permission, operation string) bool {
	switch permission {
	case PermissionQueryCreate:
		return operation == OperationQueryCreate
	case PermissionQueryRead:
		return operation == OperationQueryGet || operation == OperationQueryGetPublished ||
			operation == OperationQueryList
	case PermissionQueryUpdate:
		return operation == OperationQueryUpdate
	case PermissionQueryDelete:
		return operation == OperationQueryDelete
	case PermissionQueryPublish:
		return operation == OperationQueryPublish || operation == OperationQueryRevalidate
	case PermissionQueryExecute:
		return operation == OperationQueryExecute
	case PermissionDashboardCreate:
		return operation == OperationDashboardCreate
	case PermissionDashboardRead:
		return operation == OperationDashboardGet || operation == OperationDashboardList ||
			operation == OperationDashboardRead || operation == OperationDashboardSetDefault
	case PermissionDashboardUpdate:
		return operation == OperationDashboardUpdate || operation == OperationDashboardGrantACL ||
			operation == OperationDashboardRevokeACL
	case PermissionDashboardDelete:
		return operation == OperationDashboardDelete
	case PermissionDashboardPublish:
		return operation == OperationDashboardPublish
	case PermissionDashboardExport:
		return operation == OperationDashboardExportCreate ||
			operation == OperationDashboardExportGet ||
			operation == OperationDashboardExportDelete ||
			operation == OperationDashboardExportContent
	case PermissionPublicationRead:
		return operation == OperationPublicationCreate || operation == OperationPublicationGet ||
			operation == OperationPublicationList || operation == OperationPublicationSubmit ||
			operation == OperationPublicationApprove || operation == OperationPublicationActivate
	case PermissionPublicationRevoke:
		return operation == OperationPublicationRevoke
	case PermissionPublicationPublishPublic, PermissionPublicationPublishCapability:
		return operation == OperationPublicationCreate
	default:
		return false
	}
}
