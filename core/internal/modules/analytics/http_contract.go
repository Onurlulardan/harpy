package analytics

import analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"

const (
	PermissionQueryCreate                  = analyticshttp.PermissionQueryCreate
	PermissionQueryRead                    = analyticshttp.PermissionQueryRead
	PermissionQueryUpdate                  = analyticshttp.PermissionQueryUpdate
	PermissionQueryDelete                  = analyticshttp.PermissionQueryDelete
	PermissionQueryPublish                 = analyticshttp.PermissionQueryPublish
	PermissionQueryExecute                 = analyticshttp.PermissionQueryExecute
	PermissionDashboardCreate              = analyticshttp.PermissionDashboardCreate
	PermissionDashboardRead                = analyticshttp.PermissionDashboardRead
	PermissionDashboardUpdate              = analyticshttp.PermissionDashboardUpdate
	PermissionDashboardDelete              = analyticshttp.PermissionDashboardDelete
	PermissionDashboardPublish             = analyticshttp.PermissionDashboardPublish
	PermissionDashboardExport              = analyticshttp.PermissionDashboardExport
	PermissionPublicationRead              = analyticshttp.PermissionPublicationRead
	PermissionPublicationRevoke            = analyticshttp.PermissionPublicationRevoke
	PermissionPublicationPublishPublic     = analyticshttp.PermissionPublicationPublishPublic
	PermissionPublicationPublishCapability = analyticshttp.PermissionPublicationPublishCapability
)

const (
	OperationQueryCreate            = analyticshttp.OperationQueryCreate
	OperationQueryGet               = analyticshttp.OperationQueryGet
	OperationQueryList              = analyticshttp.OperationQueryList
	OperationQueryUpdate            = analyticshttp.OperationQueryUpdate
	OperationQueryDelete            = analyticshttp.OperationQueryDelete
	OperationQueryPublish           = analyticshttp.OperationQueryPublish
	OperationQueryRevalidate        = analyticshttp.OperationQueryRevalidate
	OperationQueryExecute           = analyticshttp.OperationQueryExecute
	OperationDashboardCreate        = analyticshttp.OperationDashboardCreate
	OperationDashboardGet           = analyticshttp.OperationDashboardGet
	OperationDashboardList          = analyticshttp.OperationDashboardList
	OperationDashboardUpdate        = analyticshttp.OperationDashboardUpdate
	OperationDashboardDelete        = analyticshttp.OperationDashboardDelete
	OperationDashboardPublish       = analyticshttp.OperationDashboardPublish
	OperationDashboardRead          = analyticshttp.OperationDashboardRead
	OperationDashboardGrantACL      = analyticshttp.OperationDashboardGrantACL
	OperationDashboardRevokeACL     = analyticshttp.OperationDashboardRevokeACL
	OperationDashboardSetDefault    = analyticshttp.OperationDashboardSetDefault
	OperationDashboardExportCreate  = analyticshttp.OperationDashboardExportCreate
	OperationDashboardExportGet     = analyticshttp.OperationDashboardExportGet
	OperationDashboardExportDelete  = analyticshttp.OperationDashboardExportDelete
	OperationDashboardExportContent = analyticshttp.OperationDashboardExportContent
	OperationPublicationCreate      = analyticshttp.OperationPublicationCreate
	OperationPublicationGet         = analyticshttp.OperationPublicationGet
	OperationPublicationList        = analyticshttp.OperationPublicationList
	OperationPublicationSubmit      = analyticshttp.OperationPublicationSubmit
	OperationPublicationApprove     = analyticshttp.OperationPublicationApprove
	OperationPublicationActivate    = analyticshttp.OperationPublicationActivate
	OperationPublicationRevoke      = analyticshttp.OperationPublicationRevoke
)

var (
	ErrHTTPAuthentication = analyticshttp.ErrAuthentication
	ErrHTTPForbidden      = analyticshttp.ErrForbidden
	ErrHTTPInvalid        = analyticshttp.ErrInvalid
	ErrHTTPNotFound       = analyticshttp.ErrNotFound
	ErrHTTPConflict       = analyticshttp.ErrConflict
	ErrHTTPRateLimited    = analyticshttp.ErrRateLimited
	ErrHTTPUnavailable    = analyticshttp.ErrUnavailable
)
