package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (reports *platformReports) inCommandSnapshot(
	ctx context.Context,
	access billingapi.Access,
	operation string,
	action func(context.Context, platformtx.GlobalScope) error,
) error {
	if reports == nil || reports.global == nil || ctx == nil || ctx.Err() != nil ||
		!access.ValidFor(billingapi.PermissionPlatformBillingManage) ||
		access.Scope() != billingapi.AccessScopeGlobal {
		return billingapi.ErrForbidden
	}
	_, capability, found := access.GlobalCommandCapabilitiesFor(operation)
	if !found {
		return billingapi.ErrAuthorityUnavailable
	}
	var actionErr error
	err := reports.global.Snapshot(
		ctx, capability,
		func(scoped context.Context, scope platformtx.GlobalScope) error {
			actionErr = action(scoped, scope)
			return actionErr
		},
	)
	if actionErr != nil {
		return actionErr
	}
	if err != nil {
		return billingapi.ErrAuthorityUnavailable
	}
	return nil
}
