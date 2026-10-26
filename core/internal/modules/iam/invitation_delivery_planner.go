package iam

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

var invitationDeliveryEffectNamespace = uuid.MustParse("9fb18116-a96a-4afc-96e1-e27998a6031c")

type invitationDeliveryPlanner struct{}

func (invitationDeliveryPlanner) Plan(
	_ context.Context,
	claim asyncpostgres.JobClaim,
) (asyncpostgres.ExternalEffectRequestSpec, error) {
	tenantID, tenantScoped := claim.TenantID()
	reference, err := deliveryartifactv1.ParseReference(claim.PayloadReference())
	digest := sha256.Sum256([]byte(
		"iam.invitation.delivery.v1\x00" + claim.PayloadReference(),
	))
	effectKey, hasEffectKey := claim.EffectIdempotencyKey()
	if err != nil || !tenantScoped || reference.TenantID() != tenantID ||
		claim.Kind() != InvitationDeliveryJobKind || claim.PayloadRevision() != 1 ||
		claim.PayloadSHA256() != hex.EncodeToString(digest[:]) || !hasEffectKey ||
		effectKey != "iam.invitation.delivery:"+claim.PayloadReference() {
		return asyncpostgres.ExternalEffectRequestSpec{}, fmt.Errorf("invitation delivery claim is invalid")
	}
	attemptID := claim.AttemptID()
	effectID := uuid.NewSHA1(invitationDeliveryEffectNamespace, attemptID[:])
	return asyncpostgres.NewExternalEffectRequestSpec(
		effectID, "iam.smtp", "send_invitation", claim.JobID().String(), claim.JobID(),
	)
}
