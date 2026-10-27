package analyticshttp

import (
	"context"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	queryexecutionapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/queryexecution"
	"github.com/google/uuid"
)

type Application interface {
	CreateQuery(context.Context, Access, queryapp.CreateRequest) (queryapp.Snapshot, error)
	GetQuery(context.Context, Access, uuid.UUID) (queryapp.Snapshot, error)
	GetPublishedQuery(context.Context, Access, uuid.UUID) (queryapp.PublishedSnapshot, error)
	ListQueries(context.Context, Access, int, *uuid.UUID) (queryapp.Page, error)
	UpdateQuery(context.Context, Access, queryapp.UpdateRequest) (queryapp.Snapshot, error)
	DeleteQuery(context.Context, Access, queryapp.DeleteRequest) (queryapp.Snapshot, error)
	PublishQuery(context.Context, Access, queryapp.PublishRequest) (queryapp.Snapshot, error)
	RevalidateQuery(context.Context, Access, queryapp.PublishRequest) (queryapp.Snapshot, error)
	ExecuteQuery(context.Context, Access, queryapp.ExecuteRequest) (queryapp.ExecutionResult, error)
	ExecuteQueryBatch(context.Context, Access, queryexecutionapp.Request) (queryexecutionapp.Result, error)
	CreateDashboard(context.Context, Access, dashboardapp.CreateRequest) (dashboardapp.Snapshot, error)
	GetDashboard(context.Context, Access, uuid.UUID) (dashboardapp.Snapshot, error)
	GetDefaultDashboard(context.Context, Access, string) (dashboardapp.Snapshot, error)
	ListDashboards(context.Context, Access, int, *uuid.UUID) (dashboardapp.Page, error)
	UpdateDashboard(context.Context, Access, dashboardapp.UpdateRequest) (dashboardapp.Snapshot, error)
	DeleteDashboard(context.Context, Access, dashboardapp.DeleteRequest) (dashboardapp.Snapshot, error)
	PublishDashboard(context.Context, Access, dashboardapp.PublishRequest) (dashboardapp.Snapshot, error)
	GrantDashboardACL(context.Context, Access, dashboardapp.GrantACLRequest) (dashboardapp.ACL, error)
	RevokeDashboardACL(context.Context, Access, dashboardapp.RevokeACLRequest) error
	SetDefaultDashboard(context.Context, Access, dashboardapp.SetDefaultRequest) error
	CreatePublication(context.Context, Access, publicationapp.CreateRequest) (publicationapp.Snapshot, error)
	GetPublication(context.Context, Access, uuid.UUID) (publicationapp.Snapshot, error)
	ListPublications(context.Context, Access, int, *uuid.UUID) (publicationapp.Page, error)
	SubmitPublication(context.Context, Access, publicationapp.TransitionRequest) (publicationapp.Snapshot, error)
	ApprovePublication(context.Context, Access, publicationapp.TransitionRequest) (publicationapp.Snapshot, error)
	ActivatePublication(context.Context, Access, publicationapp.ActivateRequest) (publicationapp.Activation, error)
	RevokePublication(context.Context, Access, publicationapp.TransitionRequest) (publicationapp.Snapshot, error)
}

type PublicApplication interface {
	Execute(context.Context, publicationapp.PublicRequest) (publicationapp.PublicResult, error)
}

type LegacyApplication interface {
	AuthorizeLegacy(context.Context, compatibilityapp.LegacyKey) (compatibilityapp.LegacyMapping, error)
}
