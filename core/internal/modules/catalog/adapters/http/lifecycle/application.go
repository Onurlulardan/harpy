package lifecyclehttp

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/google/uuid"
)

type Application interface {
	ListInstallations(context.Context, cataloghttp.Access, int, *uuid.UUID) (
		[]cataloghttp.Installation, *uuid.UUID, error)
	Install(context.Context, cataloghttp.Access, uuid.UUID, string, string) (
		cataloghttp.Operation, error)
	GetInstallation(context.Context, cataloghttp.Access, uuid.UUID) (
		cataloghttp.Installation, error)
	Upgrade(context.Context, cataloghttp.Access, uuid.UUID, string, string) (
		cataloghttp.Operation, error)
	Disable(context.Context, cataloghttp.Access, uuid.UUID, string) (
		cataloghttp.Installation, error)
	Enable(context.Context, cataloghttp.Access, uuid.UUID, string) (
		cataloghttp.Installation, error)
	Restore(context.Context, cataloghttp.Access, uuid.UUID, string) (
		cataloghttp.Operation, error)
	Retire(context.Context, cataloghttp.Access, uuid.UUID, string) (
		cataloghttp.Operation, error)
	Purge(context.Context, cataloghttp.Access, cataloghttp.PurgeInput) (
		cataloghttp.Operation, error)
	ApprovePurge(context.Context, cataloghttp.Access, cataloghttp.PurgeApprovalInput) (
		cataloghttp.PurgeApproval, error)
	GetOperation(context.Context, cataloghttp.Access, uuid.UUID) (cataloghttp.Operation, error)
	CancelOperation(context.Context, cataloghttp.Access, uuid.UUID, string) (
		cataloghttp.Operation, error)
	RetryOperation(context.Context, cataloghttp.Access, uuid.UUID, string) (
		cataloghttp.Operation, error)
}

type DemoApplication interface {
	GetDemoStatus(context.Context, cataloghttp.Access, uuid.UUID) (
		cataloghttp.DemoStatus, error)
	PlanDemoRemoval(context.Context, cataloghttp.Access, uuid.UUID) (
		cataloghttp.DemoRemovalPlan, error)
	InstallDemo(context.Context, cataloghttp.Access, uuid.UUID, string) (
		cataloghttp.Operation, error)
	RemoveDemo(context.Context, cataloghttp.Access, uuid.UUID, string, string) (
		cataloghttp.Operation, error)
}
