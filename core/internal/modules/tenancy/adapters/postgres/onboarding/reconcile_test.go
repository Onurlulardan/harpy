package onboarding

import (
	"context"
	"strings"
	"testing"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
	"github.com/google/uuid"
)

func TestReconcileUsesPersistedEvidenceRoutine(t *testing.T) {
	operationID := uuid.New()
	storage := &postgresStore{execute: noopExecute, queryRow: func(
		_ context.Context, statement string, arguments ...any,
	) Row {
		if !strings.Contains(statement, "platform.tenancy_onboarding_reconcile_v1") ||
			len(arguments) != 5 {
			t.Fatalf("reconcile query = %q %#v", statement, arguments)
		}
		return transitionRow(operationID, tenantapi.OnboardingMigrationReady, 5)
	}}
	transition, err := storage.Reconcile(context.Background(), onboardingapp.ReconcileRequest{
		OperationID: operationID, ReceiptID: uuid.New(),
		ReceiptSHA256: strings.Repeat("a", 64), EvidenceSHA256: strings.Repeat("b", 64),
		ExpectedRevision: 4})
	if err != nil || transition.State != tenantapi.OnboardingMigrationReady {
		t.Fatalf("transition = %#v, %v", transition, err)
	}
}
