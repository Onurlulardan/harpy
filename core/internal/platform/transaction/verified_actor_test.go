package transaction

import (
	"testing"

	"github.com/google/uuid"
)

func TestActorIssuerProducesOpaqueTenantCapability(t *testing.T) {
	issuer := NewActorIssuer()
	actorID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	tenantID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	actor, err := issuer.Issue(actorID, tenantID)
	if err != nil || !actor.valid() || actor.actorID != actorID || actor.tenantID != tenantID {
		t.Fatalf("unexpected actor capability: %+v %v", actor, err)
	}
	if _, err := (ActorIssuer{}).Issue(actorID, tenantID); err == nil {
		t.Fatal("zero issuer minted an actor capability")
	}
}
