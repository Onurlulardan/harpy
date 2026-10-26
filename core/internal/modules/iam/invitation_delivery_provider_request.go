package iam

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
	"github.com/google/uuid"
)

func (provider *invitationDeliveryProvider) parseRequest(
	request effectsv1.Request,
) (invitationRequest, error) {
	if provider == nil || request.Validate() != nil || request.Provider != "iam.smtp" ||
		request.Operation != "send_invitation" || request.EffectKind != InvitationDeliveryJobKind ||
		request.EffectRevision != 1 || request.TenantID == nil ||
		request.Authority.Domain != "iam" {
		return invitationRequest{}, ErrInvitationDeliveryRuntime
	}
	tenantID, err := uuid.Parse(*request.TenantID)
	reference, referenceErr := deliveryartifactv1.ParseReference(request.PayloadReference)
	digest := sha256.Sum256([]byte(
		"iam.invitation.delivery.v1\x00" + request.PayloadReference,
	))
	requestedAt, timeErr := time.Parse(time.RFC3339Nano, request.RequestedAt)
	messageID, messageErr := provider.smtp.StableMessageID(request.IdempotencyKey)
	if err != nil || referenceErr != nil || timeErr != nil || messageErr != nil ||
		tenantID == uuid.Nil || reference.TenantID() != tenantID ||
		request.PayloadSHA256 != hex.EncodeToString(digest[:]) ||
		request.IdempotencyKey != "iam.invitation.delivery:"+request.PayloadReference {
		return invitationRequest{}, ErrInvitationDeliveryRuntime
	}
	effectID := uuid.MustParse(request.EffectID)
	return invitationRequest{
		request: request, tenantID: tenantID, requestedAt: requestedAt,
		messageID: messageID,
		reference: invitationdeliveryapp.EffectReference{
			TenantID: tenantID, ArtifactRef: request.PayloadReference,
			EffectID: effectID, EffectKey: request.IdempotencyKey,
			OccurredAt: provider.now().UTC(),
		},
	}, nil
}
