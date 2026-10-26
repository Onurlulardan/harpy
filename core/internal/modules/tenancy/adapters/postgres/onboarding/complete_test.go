package onboarding

import (
	"context"
	"strings"
	"testing"

	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
	"github.com/google/uuid"
)

func TestCompleteActivationUsesSameTransactionRoutine(t *testing.T) {
	called := false
	storage := &postgresStore{queryRow: func(context.Context, string, ...any) Row { return nil },
		execute: func(_ context.Context, statement string, arguments ...any) error {
			called = strings.Contains(statement, "platform.tenancy_onboarding_complete_v2") &&
				len(arguments) == 5
			return nil
		}}
	err := storage.completeActivation(context.Background(), onboardingapp.ActivationCompletion{
		OperationID: uuid.New(), ReceiptID: uuid.New(), ReceiptSHA256: strings.Repeat("a", 64),
		EvidenceSHA256: strings.Repeat("b", 64), ExpectedRevision: 4})
	if err != nil || !called {
		t.Fatalf("completion called=%t err=%v", called, err)
	}
}
