package iam

import (
	"context"
	"errors"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/gcpkms"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/smtp"
	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func (provider *invitationDeliveryProvider) Execute(
	ctx context.Context,
	request effectsv1.Request,
) (result effectexecution.Result) {
	parsed, err := provider.parseRequest(request)
	if err != nil || ctx == nil || ctx.Err() != nil {
		return effectexecution.Unknown()
	}
	claimed := false
	defer func() {
		if recover() != nil {
			if claimed {
				provider.persistUnknown(ctx, parsed.reference)
			}
			result = effectexecution.Unknown()
		}
	}()
	decision, tenantSlug, err := provider.state.Claim(ctx, invitationdeliveryapp.ClaimCommand{
		TenantID: parsed.tenantID, ArtifactRef: request.PayloadReference,
		EffectID: parsed.reference.EffectID, EffectKey: request.IdempotencyKey,
		MessageID: parsed.messageID, ClaimedAt: parsed.reference.OccurredAt,
	})
	if err != nil {
		if errors.Is(err, invitationdeliveryapp.ErrArtifactUnavailable) {
			return provider.failed(parsed, "DELIVERY_ARTIFACT_UNAVAILABLE", effectsv1.ErrorPermanent)
		}
		return effectexecution.Unknown()
	}
	if decision.Replay != nil {
		return provider.receiptForStored(parsed, *decision.Replay)
	}
	if decision.Held || decision.Artifact == nil {
		return effectexecution.Unknown()
	}
	claimed = true
	credential, err := provider.opener.Open(
		ctx, parsed.tenantID, decision.Artifact.Envelope,
	)
	if err != nil {
		if errors.Is(err, gcpkms.ErrUnavailable) {
			return provider.releaseThenFailed(ctx, parsed, "KMS_UNAVAILABLE")
		}
		return provider.finalizePermanent(
			ctx, parsed, "DELIVERY_ARTIFACT_INVALID",
			deliveryFailureDigest("DELIVERY_ARTIFACT_INVALID", parsed.reference.ArtifactRef),
		)
	}
	defer credential.Destroy()
	var delivered smtp.Result
	err = credential.RevealOnce(func(token []byte) error {
		delivered = provider.smtp.Send(ctx, smtp.Message{
			Recipient: decision.Artifact.Recipient, MessageID: parsed.messageID,
			TenantSlug: tenantSlug, Token: token, Date: parsed.requestedAt,
		})
		return nil
	})
	if err != nil {
		provider.persistUnknown(ctx, parsed.reference)
		return effectexecution.Unknown()
	}
	return provider.handleSMTPResult(ctx, parsed, delivered)
}
