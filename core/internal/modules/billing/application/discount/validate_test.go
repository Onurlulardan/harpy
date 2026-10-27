package discount

import (
	"context"
	"errors"
	"testing"
	"time"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type rejectingEligibilityStore struct{ called bool }

func (store *rejectingEligibilityStore) FindEligible(
	context.Context, EligibilityLookup,
) (Eligibility, error) {
	store.called = true
	return Eligibility{}, ErrQueryStorage
}

func TestValidateRejectsUnboundScopeBeforeStorage(t *testing.T) {
	store := &rejectingEligibilityStore{}
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	queries, err := NewQueries(store, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	code, _ := discountdomain.NormalizeCode("launch")
	_, err = queries.Validate(context.Background(), platformtx.TenantScope{}, ValidateRequest{
		Code: string(code), OfferingVersionID: mustV7(t),
	})
	if !errors.Is(err, ErrQueryScope) || store.called {
		t.Fatalf("validate error=%v called=%t", err, store.called)
	}
}

func TestNewEligibilityRequiresMatchingTypedPolicy(t *testing.T) {
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	code, _ := discountdomain.NormalizeCode("launch")
	rate, _ := money.NewBasisPoints(1000)
	policy, err := discountdomain.NewBasisPointPolicy(
		code, rate, discountdomain.OnceDuration(), 1, nil, nil, now, nil, true, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	value, err := NewEligibility(EligibilityInput{
		TenantID: uuid.New(), DiscountID: mustV7(t), DiscountVersionID: mustV7(t),
		SubscriptionID: mustV7(t), OfferingVersionID: mustV7(t),
		PriceVersionID: mustV7(t), Revision: 1, Version: 1,
		Currency: "USD", Policy: policy,
	})
	if err != nil || value.Policy().Code() != code || value.Currency() != "USD" {
		t.Fatalf("eligibility=%+v error=%v", value, err)
	}
}
