package analyticshttp

import httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"

const (
	OperationQueryCreate            = httptransport.OperationQueryCreate
	OperationQueryGet               = httptransport.OperationQueryGet
	OperationQueryGetPublished      = httptransport.OperationQueryGetPublished
	OperationQueryList              = httptransport.OperationQueryList
	OperationQueryUpdate            = httptransport.OperationQueryUpdate
	OperationQueryDelete            = httptransport.OperationQueryDelete
	OperationQueryPublish           = httptransport.OperationQueryPublish
	OperationQueryRevalidate        = httptransport.OperationQueryRevalidate
	OperationQueryExecute           = httptransport.OperationQueryExecute
	OperationDashboardCreate        = httptransport.OperationDashboardCreate
	OperationDashboardGet           = httptransport.OperationDashboardGet
	OperationDashboardList          = httptransport.OperationDashboardList
	OperationDashboardUpdate        = httptransport.OperationDashboardUpdate
	OperationDashboardDelete        = httptransport.OperationDashboardDelete
	OperationDashboardPublish       = httptransport.OperationDashboardPublish
	OperationDashboardRead          = httptransport.OperationDashboardRead
	OperationDashboardGrantACL      = httptransport.OperationDashboardGrantACL
	OperationDashboardRevokeACL     = httptransport.OperationDashboardRevokeACL
	OperationDashboardSetDefault    = httptransport.OperationDashboardSetDefault
	OperationDashboardExportCreate  = httptransport.OperationDashboardExportCreate
	OperationDashboardExportGet     = httptransport.OperationDashboardExportGet
	OperationDashboardExportDelete  = httptransport.OperationDashboardExportDelete
	OperationDashboardExportContent = httptransport.OperationDashboardExportContent
	OperationPublicationCreate      = httptransport.OperationPublicationCreate
	OperationPublicationGet         = httptransport.OperationPublicationGet
	OperationPublicationList        = httptransport.OperationPublicationList
	OperationPublicationSubmit      = httptransport.OperationPublicationSubmit
	OperationPublicationApprove     = httptransport.OperationPublicationApprove
	OperationPublicationActivate    = httptransport.OperationPublicationActivate
	OperationPublicationRevoke      = httptransport.OperationPublicationRevoke
)

const (
	PermissionQueryCreate                  = httptransport.PermissionQueryCreate
	PermissionQueryRead                    = httptransport.PermissionQueryRead
	PermissionQueryUpdate                  = httptransport.PermissionQueryUpdate
	PermissionQueryDelete                  = httptransport.PermissionQueryDelete
	PermissionQueryPublish                 = httptransport.PermissionQueryPublish
	PermissionQueryExecute                 = httptransport.PermissionQueryExecute
	PermissionDashboardCreate              = httptransport.PermissionDashboardCreate
	PermissionDashboardRead                = httptransport.PermissionDashboardRead
	PermissionDashboardUpdate              = httptransport.PermissionDashboardUpdate
	PermissionDashboardDelete              = httptransport.PermissionDashboardDelete
	PermissionDashboardPublish             = httptransport.PermissionDashboardPublish
	PermissionDashboardExport              = httptransport.PermissionDashboardExport
	PermissionPublicationRead              = httptransport.PermissionPublicationRead
	PermissionPublicationRevoke            = httptransport.PermissionPublicationRevoke
	PermissionPublicationPublishPublic     = httptransport.PermissionPublicationPublishPublic
	PermissionPublicationPublishCapability = httptransport.PermissionPublicationPublishCapability
)
