package onboarding

import (
	"context"
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func TestRecordReceiptUsesAppendOnlyRoutine(t *testing.T) {
	operationID := uuid.New()
	storage := &postgresStore{execute: noopExecute, queryRow: func(
		_ context.Context, statement string, arguments ...any,
	) Row {
		if !strings.Contains(statement, "platform.tenancy_onboarding_record_v1") || len(arguments) != 8 {
			t.Fatalf("record query = %q %#v", statement, arguments)
		}
		return transitionRow(operationID, tenantapi.OnboardingMigrationReady, 3)
	}}
	now := time.Now().UTC().Truncate(time.Microsecond)
	transition, err := storage.RecordReceipt(context.Background(), tenantapi.OnboardingReceipt{
		ReceiptID: uuid.New(), OperationID: operationID, AttemptID: uuid.New(),
		Kind: tenantapi.ReceiptMigrationCompleted, SHA256: strings.Repeat("a", 64),
		EvidenceSHA256: strings.Repeat("b", 64), OccurredAt: now}, 2)
	if err != nil || transition.State != tenantapi.OnboardingMigrationReady {
		t.Fatalf("transition = %#v, %v", transition, err)
	}
}
