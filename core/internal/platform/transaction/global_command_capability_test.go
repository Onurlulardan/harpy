package transaction

import (
	"testing"

	"github.com/google/uuid"
)

func TestGlobalCommandCapabilityRequiresAuthenticatedFreshBillingGrant(t *testing.T) {
	issuer := NewGlobalCommandCapabilityIssuer()
	actorID := uuid.New()
	capability, err := issuer.Issue(
		actorID, "billing.catalog.publish", "platform.billing.catalog.manage",
		9, 7, 5, 3,
	)
	if err != nil || !capability.valid() || capability.actorID != actorID {
		t.Fatalf("global command capability = %#v, %v", capability, err)
	}
	cases := []struct {
		actor      uuid.UUID
		permission string
		clock      int64
		version    int64
		group      int64
		domain     int64
	}{
		{uuid.Nil, "platform.billing.catalog.manage", 9, 7, 5, 3},
		{actorID, "billing.manage", 9, 7, 5, 3},
		{actorID, "platform.billing.*", 9, 7, 5, 3},
		{actorID, "platform.billing.catalog.manage", 0, 7, 5, 3},
		{actorID, "platform.billing.catalog.manage", 9, 0, 5, 3},
		{actorID, "platform.billing.catalog.manage", 9, 7, 0, 3},
		{actorID, "platform.billing.catalog.manage", 9, 7, 5, 0},
	}
	for index, value := range cases {
		if _, err := issuer.Issue(
			value.actor, "billing.catalog.publish", value.permission,
			value.clock, value.version, value.group, value.domain,
		); err == nil {
			t.Fatalf("invalid global command capability %d was accepted", index)
		}
	}
}
