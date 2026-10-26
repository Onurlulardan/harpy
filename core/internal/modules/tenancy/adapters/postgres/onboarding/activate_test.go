package onboarding

import (
	"context"
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
	"github.com/google/uuid"
)

func TestActivateUsesDatabaseClockCASAndCompletionOnOneTransaction(t *testing.T) {
	operationID, tenantID, receiptID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	completed := false
	storage := &postgresStore{execute: func(
		_ context.Context, statement string, arguments ...any,
	) error {
		completed = strings.Contains(statement, "platform.tenancy_onboarding_complete_v2") &&
			len(arguments) == 5
		return nil
	}, queryRow: func(
		_ context.Context, statement string, arguments ...any,
	) Row {
		if !strings.Contains(statement, "platform.tenancy_onboarding_activate_v3") ||
			len(arguments) != 3 {
			t.Fatalf("activate query = %q %#v", statement, arguments)
		}
		return rowFixture(func(destinations ...any) error {
			*destinations[0].(*uuid.UUID), *destinations[1].(*uuid.UUID) = operationID, tenantID
			*destinations[2].(*string) = string(tenantapi.OnboardingEnvironmentInitialize)
			*destinations[3].(*string) = strings.Repeat("a", 64)
			*destinations[4].(*uuid.UUID), *destinations[5].(*string) = receiptID, strings.Repeat("b", 64)
			*destinations[6].(**string) = nil
			*destinations[7].(*int64) = 4
			*destinations[8].(*time.Time), *destinations[9].(*time.Time) = now, now.Add(time.Minute)
			return nil
		})
	}}
	transaction := &activationTransaction{store: storage,
		bind: func(onboardingapp.ActivationOwner) (onboardingapp.ActivationOwnerExecutor, error) {
			return onboardingapp.ActivationOwnerExecutor{}, nil
		}, bindSchema: func(context.Context, string) error { return nil }}
	transaction.open.Store(true)
	err := transaction.Activate(context.Background(), onboardingapp.ActivationRequest{
		OperationID: operationID, LeaseTokenSHA256: strings.Repeat("c", 64), ExpectedRevision: 3},
		onboardingapp.ActivationCompletion{OperationID: operationID, ReceiptID: uuid.New(),
			ReceiptSHA256: strings.Repeat("d", 64), EvidenceSHA256: strings.Repeat("e", 64)},
		func(claim tenantapi.ClaimedActivation, _ onboardingapp.ActivationResources) error {
			if claim.DatabaseNow != now || claim.LeaseUntil != now.Add(time.Minute) {
				t.Fatalf("claim = %#v", claim)
			}
			return nil
		})
	if err != nil || !completed {
		t.Fatalf("activation completed=%t err=%v", completed, err)
	}
}
