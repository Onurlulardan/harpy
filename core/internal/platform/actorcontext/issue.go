package actorcontext

import (
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func Issue(actorID, tenantID uuid.UUID) (platformtx.VerifiedActor, error) {
	return platformtx.NewActorIssuer().Issue(actorID, tenantID)
}
