package dashboardhttp

import (
	"context"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	"github.com/google/uuid"
)

type Application interface {
	CreateDashboard(context.Context, httptransport.Access, dashboardapp.CreateRequest) (dashboardapp.Snapshot, error)
	GetDashboard(context.Context, httptransport.Access, uuid.UUID) (dashboardapp.Snapshot, error)
	GetDefaultDashboard(context.Context, httptransport.Access, string) (dashboardapp.Snapshot, error)
	ListDashboards(context.Context, httptransport.Access, int, *uuid.UUID) (dashboardapp.Page, error)
	UpdateDashboard(context.Context, httptransport.Access, dashboardapp.UpdateRequest) (dashboardapp.Snapshot, error)
	DeleteDashboard(context.Context, httptransport.Access, dashboardapp.DeleteRequest) (dashboardapp.Snapshot, error)
	PublishDashboard(context.Context, httptransport.Access, dashboardapp.PublishRequest) (dashboardapp.Snapshot, error)
	GrantDashboardACL(context.Context, httptransport.Access, dashboardapp.GrantACLRequest) (dashboardapp.ACL, error)
	RevokeDashboardACL(context.Context, httptransport.Access, dashboardapp.RevokeACLRequest) error
	SetDefaultDashboard(context.Context, httptransport.Access, dashboardapp.SetDefaultRequest) error
}
