package onboarding

import (
	"context"
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func TestReceiveAuthorityAckUsesScopedCASRoutine(t *testing.T) {
	operationID := uuid.New()
	storage := &postgresStore{execute: noopExecute, queryRow: func(
		_ context.Context, statement string, arguments ...any,
	) Row {
		if !strings.Contains(statement, "platform.tenancy_onboarding_authority_ack_v1") ||
			len(arguments) != 12 {
			t.Fatalf("authority ack query = %q %#v", statement, arguments)
		}
		return transitionRow(operationID, tenantapi.OnboardingAuthorityAcknowledged, 4)
	}}
	ack := tenantapi.AuthorityAcknowledgement{OperationID: operationID,
		RegistryBindingID: uuid.New(), ScopeKind: tenantapi.AuthorityScopeEnvironment,
		ScopeID: uuid.New(), Generation: 1, StateSHA256: strings.Repeat("a", 64),
		ProposalSHA256: strings.Repeat("b", 64), AckSHA256: strings.Repeat("c", 64),
		RegistrySequence: 1, VerifiedAt: time.Now().UTC().Truncate(time.Microsecond)}
	transition, err := storage.ReceiveAuthorityAck(context.Background(), ack, 3)
	if err != nil || transition.OperationID != operationID {
		t.Fatalf("authority transition = %#v, %v", transition, err)
	}
}
