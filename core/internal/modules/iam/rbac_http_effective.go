package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) Effective(
	ctx context.Context,
	claims accesstoken.Claims,
	input RBACEffectiveRead,
) (RBACEffectivePage, error) {
	if !rbacEffectiveReadValid(input) {
		return RBACEffectivePage{}, ErrRBACHTTPInvalid
	}
	page := RBACEffectivePage{}
	err := application.inSnapshot(
		ctx, claims, "iam.permissions", "read",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			values, readErr := application.runtime.assignments.Effective(
				scoped, scope, rbacdomain.EffectivePermissionQuery{
					Target: input.Target, TargetID: rbacdomain.ID(input.TargetID),
					Limit:    input.Limit,
					Evidence: rbacQueryEvidence(claims, permissionID),
				},
			)
			if readErr != nil {
				return readErr
			}
			page.Items = make([]RBACEffectiveView, 0, len(values))
			for _, value := range values {
				page.Items = append(page.Items, rbacEffectiveView(value))
			}
			return nil
		},
	)
	if err != nil {
		return RBACEffectivePage{}, err
	}
	return page, nil
}
