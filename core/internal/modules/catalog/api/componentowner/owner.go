package componentowner

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Owner interface {
	Readiness(context.Context, platformtx.TenantScope, Component) (Readiness, error)
	Stage(context.Context, platformtx.TenantScope, Command) (Receipt, error)
	Cleanup(context.Context, platformtx.TenantScope, Command) (Receipt, error)
	Activate(context.Context, platformtx.TenantScope, Command) (Receipt, error)
	Disable(context.Context, platformtx.TenantScope, Command) (Receipt, error)
	Enable(context.Context, platformtx.TenantScope, Command) (Receipt, error)
	Retire(context.Context, platformtx.TenantScope, Command) (Receipt, error)
	Restore(context.Context, platformtx.TenantScope, Command) (Receipt, error)
	PurgeBlockers(context.Context, platformtx.TenantScope, Command) (BlockerReceipt, error)
	Purge(context.Context, platformtx.TenantScope, Command) (Receipt, error)
}

type LifecycleOwner interface {
	RetireComponent(
		context.Context, platformtx.TenantScope, LifecycleCommand,
	) (OwnerResult, error)
	RestoreComponent(
		context.Context, platformtx.TenantScope, LifecycleCommand,
	) (OwnerResult, error)
	PreflightComponentPurge(
		context.Context, platformtx.TenantScope, PurgeCommand,
	) (OwnerResult, error)
	BeginComponentPurge(
		context.Context, platformtx.TenantScope, PurgeCommand,
	) (OwnerResult, error)
	FinalizeComponentPurge(
		context.Context, platformtx.TenantScope, PurgeCommand,
	) (OwnerResult, error)
}

// Provisioner materializes a sealed release payload in the owning module.
// Catalog passes only immutable registry evidence; HTTP request bodies are
// never accepted as a component-definition authority.
type Provisioner interface {
	Target(
		context.Context,
		platformtx.TenantScope,
		Component,
	) (Component, error)
	Provision(
		context.Context,
		platformtx.TenantScope,
		ProvisionRequest,
	) (ProvisionReceipt, error)
}

type Directory interface {
	Resolve(owner string) (Owner, error)
}
