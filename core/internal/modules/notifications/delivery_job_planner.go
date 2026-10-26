package notifications

import (
	"context"
	"fmt"

	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

var deliveryEffectNamespace = uuid.MustParse("91cbef78-b125-43ba-8b50-6e720498e842")

type deliveryPlanner struct{ kind, provider string }

func (planner deliveryPlanner) Plan(_ context.Context,
	claim asyncpostgres.JobClaim,
) (asyncpostgres.ExternalEffectRequestSpec, error) {
	tenantID, tenantScoped := claim.TenantID()
	reference, err := parseDeliveryReference(claim.PayloadReference())
	effectKey, hasEffectKey := claim.EffectIdempotencyKey()
	wantEffect := "notifications.delivery:" + reference.deliveryID.String()
	if err != nil || !tenantScoped || reference.tenantID.String() != tenantID.String() ||
		claim.Kind() != planner.kind || claim.PayloadRevision() != 1 ||
		claim.PayloadSHA256() != deliveryReferenceDigest(claim.PayloadReference()) ||
		!hasEffectKey || effectKey != wantEffect {
		return asyncpostgres.ExternalEffectRequestSpec{},
			fmt.Errorf("notification delivery claim is invalid")
	}
	attemptID := claim.AttemptID()
	effectID := uuid.NewSHA1(deliveryEffectNamespace, attemptID[:])
	return asyncpostgres.NewExternalEffectRequestSpec(effectID, planner.provider,
		"send", claim.JobID().String(), claim.JobID())
}
