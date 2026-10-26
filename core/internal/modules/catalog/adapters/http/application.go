package cataloghttp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Application interface {
	ListPackages(context.Context, Access, int, *uuid.UUID) ([]Package, *uuid.UUID, error)
	GetPackage(context.Context, Access, uuid.UUID) (Package, error)
	ListPackageVersions(context.Context, Access, uuid.UUID) ([]PackageVersion, error)
	GetPackageVersion(context.Context, Access, uuid.UUID, string) (PackageVersion, error)
	ListPackageVersionDependencies(context.Context, Access, uuid.UUID, string) (
		[]PackageDependency, error)
	ListInstallations(context.Context, Access, int, *uuid.UUID) ([]Installation, *uuid.UUID, error)
	Install(context.Context, Access, uuid.UUID, string, string) (Operation, error)
	GetInstallation(context.Context, Access, uuid.UUID) (Installation, error)
	Upgrade(context.Context, Access, uuid.UUID, string, string) (Operation, error)
	Disable(context.Context, Access, uuid.UUID, string) (Installation, error)
	Enable(context.Context, Access, uuid.UUID, string) (Installation, error)
	Restore(context.Context, Access, uuid.UUID, string) (Operation, error)
	Retire(context.Context, Access, uuid.UUID, string) (Operation, error)
	Purge(context.Context, Access, PurgeInput) (Operation, error)
	ApprovePurge(context.Context, Access, PurgeApprovalInput) (PurgeApproval, error)
	GetOperation(context.Context, Access, uuid.UUID) (Operation, error)
	CancelOperation(context.Context, Access, uuid.UUID, string) (Operation, error)
	RetryOperation(context.Context, Access, uuid.UUID, string) (Operation, error)
	GetMenu(context.Context, Access) (Menu, error)
	GetActiveMenuConfig(context.Context, Access) (*MenuConfig, error)
	CreateMenuConfig(context.Context, Access, MenuConfigWriteInput) (MenuConfig, error)
	UpdateMenuConfig(context.Context, Access, uuid.UUID, MenuConfigWriteInput) (MenuConfig, error)
	UpdateMenuOverride(context.Context, Access, MenuOverrideInput) (MenuOverrideReceipt, error)
	ReplaceMenu(context.Context, Access, MenuReplaceInput) (Menu, error)
}

type DemoApplication interface {
	GetDemoStatus(context.Context, Access, uuid.UUID) (DemoStatus, error)
	PlanDemoRemoval(context.Context, Access, uuid.UUID) (DemoRemovalPlan, error)
	InstallDemo(context.Context, Access, uuid.UUID, string) (Operation, error)
	RemoveDemo(context.Context, Access, uuid.UUID, string, string) (Operation, error)
}

type PurgeInput struct {
	InstallationID, OperationID, BackupReceiptID, ApprovalReceiptID uuid.UUID
	StepUpAt                                                        time.Time
	Reason, ConfirmationSlug, IdempotencyKey                        string
}

type PurgeApprovalInput struct {
	InstallationID, OperationID, BackupReceiptID, RequestedBy uuid.UUID
	StepUpAt                                                  time.Time
	Reason, ConfirmationSlug, PurgeIdempotencyKey             string
	ApprovalIdempotencyKey                                    string
}

type MenuOverrideInput struct {
	ItemID           uuid.UUID
	ExpectedRevision uint64
	Label            *string
	Order            *int
	Visible          *bool
	IdempotencyKey   string
}
