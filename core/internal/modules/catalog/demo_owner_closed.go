package catalog

import (
	"context"
	"fmt"

	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type closedDemoOwner struct{}

func NewClosedDemoOwner() lifecycleapp.DemoOwner { return closedDemoOwner{} }

func (closedDemoOwner) Ready(context.Context, platformtx.TenantScope) error {
	return errClosedDemoOwner
}

func (closedDemoOwner) PlanRemoval(context.Context, platformtx.TenantScope,
	lifecycleapp.DemoOwnerCommand) (lifecycleapp.DemoOwnerResult, error) {
	return lifecycleapp.DemoOwnerResult{}, errClosedDemoOwner
}

func (closedDemoOwner) Install(context.Context, platformtx.TenantScope,
	lifecycleapp.DemoOwnerCommand) (lifecycleapp.DemoOwnerResult, error) {
	return lifecycleapp.DemoOwnerResult{}, errClosedDemoOwner
}

func (closedDemoOwner) Remove(context.Context, platformtx.TenantScope,
	lifecycleapp.DemoOwnerCommand) (lifecycleapp.DemoOwnerResult, error) {
	return lifecycleapp.DemoOwnerResult{}, errClosedDemoOwner
}

func (closedDemoOwner) Inspect(context.Context, platformtx.TenantScope,
	lifecycleapp.DemoOwnerCommand) (lifecycleapp.DemoOwnerResult, error) {
	return lifecycleapp.DemoOwnerResult{}, errClosedDemoOwner
}

var errClosedDemoOwner = fmt.Errorf("catalog demo owner is unavailable")
