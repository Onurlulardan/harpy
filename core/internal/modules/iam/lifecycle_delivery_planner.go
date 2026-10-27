package iam

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

var lifecycleDeliveryEffectNamespace = uuid.MustParse("22c30fb5-3e6f-4dd0-9819-8f0fb0c36d30")

type lifecycleDeliveryPlanner struct{}

func (lifecycleDeliveryPlanner) Plan(_ context.Context,
	claim asyncpostgres.JobClaim) (asyncpostgres.ExternalEffectRequestSpec, error) {
	tenantID, tenantScoped := claim.TenantID()
	reference, err := lifecycledeliveryv1.ParseReference(claim.PayloadReference())
	digest := sha256.Sum256([]byte(
		"iam.lifecycle.delivery.v1\x00" + claim.PayloadReference(),
	))
	effectKey, hasEffectKey := claim.EffectIdempotencyKey()
	if err != nil || !tenantScoped || reference.TenantID() != tenantID ||
		claim.Kind() != LifecycleDeliveryJobKind || claim.PayloadRevision() != 1 ||
		claim.PayloadSHA256() != hex.EncodeToString(digest[:]) || !hasEffectKey ||
		effectKey != "iam.lifecycle.delivery:"+claim.PayloadReference() {
		return asyncpostgres.ExternalEffectRequestSpec{},
			fmt.Errorf("lifecycle delivery claim is invalid")
	}
	attemptID := claim.AttemptID()
	effectID := uuid.NewSHA1(lifecycleDeliveryEffectNamespace, attemptID[:])
	return asyncpostgres.NewExternalEffectRequestSpec(effectID, "iam.smtp",
		"send_lifecycle_credential", claim.JobID().String(), claim.JobID())
}
