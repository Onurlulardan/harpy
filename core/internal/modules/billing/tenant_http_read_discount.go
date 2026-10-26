package billing

import (
	"context"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *tenantOperations) readDiscount(
	ctx context.Context,
	access billingapi.Access,
	authority generations,
	code string,
	offeringVersionID uuid.UUID,
) (billingapi.TenantDiscountSnapshot, error) {
	var result billingapi.TenantDiscountSnapshot
	err := application.inSnapshot(
		ctx, access, authority, discountapp.OperationValidate,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			queries, bindErr := application.bindings.DiscountQueries(scope)
			if bindErr != nil || queries == nil {
				return billingapi.ErrAuthorityUnavailable
			}
			var value billingapi.TenantDiscountSnapshot
			var queryErr error
			if offeringVersionID == uuid.Nil {
				value, queryErr = queries.ResolveCurrent(scoped, scope, code)
			} else {
				value, queryErr = queries.Validate(scoped, scope, billingapi.TenantDiscountLookup{
					Code: code, OfferingVersionID: offeringVersionID,
				})
			}
			if queryErr == nil {
				result = value
			}
			return queryErr
		},
	)
	if err != nil {
		return billingapi.TenantDiscountSnapshot{}, normalizeDiscountQueryError(err)
	}
	if !result.ValidForTenant(access.TenantID()) {
		return billingapi.TenantDiscountSnapshot{}, billingapi.ErrUnavailable
	}
	return result, nil
}
