package onboarding

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func TestReceiveAuthorityAckRejectsCASDrift(t *testing.T) {
	plan := planFixture(t)
	coordinator, store, _ := coordinatorFixture(t, plan)
	store.transition = tenantapi.OnboardingTransition{OperationID: uuid.New(),
		State: tenantapi.OnboardingAuthorityAcknowledged, Revision: 4,
		UpdatedAt: time.Now().UTC().Truncate(time.Microsecond)}
	ack := tenantapi.AuthorityAcknowledgement{OperationID: plan.OperationID(),
		RegistryBindingID: uuid.New(), ScopeKind: tenantapi.AuthorityScopeEnvironment,
		ScopeID: plan.EnvironmentID(), Generation: 1,
		StateSHA256: strings.Repeat("a", 64), ProposalSHA256: strings.Repeat("b", 64),
		AckSHA256: strings.Repeat("c", 64), RegistrySequence: 1,
		VerifiedAt: time.Now().UTC().Truncate(time.Microsecond)}
	_, err := coordinator.ReceiveAuthorityAck(context.Background(), ack, 3)
	if !errors.Is(err, tenantapi.ErrAuthorityCASConflict) {
		t.Fatalf("CAS drift error = %v", err)
	}
}
