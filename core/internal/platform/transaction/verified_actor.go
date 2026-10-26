package transaction

import (
	"fmt"

	"github.com/google/uuid"
)

type VerifiedActor struct {
	actorID  uuid.UUID
	tenantID uuid.UUID
	issued   bool
}

type ActorIssuer struct {
	valid bool
}

func NewActorIssuer() ActorIssuer {
	return ActorIssuer{valid: true}
}

func (issuer ActorIssuer) Issue(
	actorID uuid.UUID,
	tenantID uuid.UUID,
) (VerifiedActor, error) {
	if !issuer.valid || actorID == uuid.Nil || tenantID == uuid.Nil {
		return VerifiedActor{}, fmt.Errorf("verified actor capability is invalid")
	}
	return VerifiedActor{actorID: actorID, tenantID: tenantID, issued: true}, nil
}

func (actor VerifiedActor) valid() bool {
	return actor.issued && actor.actorID != uuid.Nil && actor.tenantID != uuid.Nil
}
