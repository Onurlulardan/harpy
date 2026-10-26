package discount

import (
	"context"
	"errors"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const OperationValidate = "billing.discount.validate"

type ValidateRequest struct {
	Code              string
	OfferingVersionID uuid.UUID
}

func (queries *Queries) Validate(
	ctx context.Context,
	scope platformtx.TenantScope,
	request ValidateRequest,
) (Eligibility, error) {
	if request.OfferingVersionID != uuid.Nil && !validV7(request.OfferingVersionID) {
		return Eligibility{}, ErrQueryInvariant
	}
	code, err := ParseCode(request.Code)
	if err != nil {
		return Eligibility{}, ErrQueryInvariant
	}
	return queries.find(ctx, scope, code, request.OfferingVersionID)
}

func (queries *Queries) find(
	ctx context.Context,
	scope platformtx.TenantScope,
	code discountdomain.Code,
	offeringVersionID uuid.UUID,
) (Eligibility, error) {
	if queries == nil || queries.store == nil || queries.clock == nil || !code.Valid() ||
		ctx == nil || ctx.Err() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) != nil ||
		scope.OperationID() != OperationValidate {
		return Eligibility{}, ErrQueryScope
	}
	now := queries.clock()
	if !validInstant(now) {
		return Eligibility{}, ErrQueryConfiguration
	}
	result, err := queries.store.FindEligible(ctx, EligibilityLookup{
		Code: code, OfferingVersionID: offeringVersionID, At: now,
	})
	if err != nil {
		return Eligibility{}, normalizeEligibilityError(err)
	}
	if result.TenantID() != scope.TenantID() || !result.valid() ||
		result.Policy().Code() != code ||
		offeringVersionID != uuid.Nil && result.OfferingVersionID() != offeringVersionID {
		return Eligibility{}, ErrQueryInvariant
	}
	return result, nil
}

func normalizeEligibilityError(err error) error {
	for _, expected := range []error{
		ErrEligibilityMissing, ErrQueryInvariant, ErrQueryDependency, ErrQueryStorage,
	} {
		if errors.Is(err, expected) {
			return expected
		}
	}
	return ErrQueryStorage
}
