package customerprovisioning

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
)

func (store *Store) Create(ctx context.Context, plan tenantapi.OnboardingPlan,
	request provisioning.AcceptedRequest) (provisioning.Operation, error) {
	if !store.usable(ctx) || !plan.IsValid() {
		return provisioning.Operation{}, provisioning.ErrInvalid
	}
	return scanOperation(
		store.queryRow(ctx, acceptSQL, acceptArguments(plan, request)...),
		store.notFound,
	)
}
