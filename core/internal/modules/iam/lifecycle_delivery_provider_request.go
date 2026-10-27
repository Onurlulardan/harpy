package iam

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
	"github.com/google/uuid"
)

func (provider *lifecycleDeliveryProvider) parseRequest(
	request effectsv1.Request) (lifecycleDeliveryRequest, error) {
	if provider == nil || request.Validate() != nil || request.Provider != "iam.smtp" ||
		request.Operation != "send_lifecycle_credential" ||
		request.EffectKind != LifecycleDeliveryJobKind || request.EffectRevision != 1 ||
		request.TenantID == nil || request.Authority.Domain != "iam" {
		return lifecycleDeliveryRequest{}, ErrLifecycleDeliveryRuntime
	}
	tenantID, tenantErr := uuid.Parse(*request.TenantID)
	effectID, effectErr := uuid.Parse(request.EffectID)
	reference, referenceErr := lifecycledeliveryv1.ParseReference(request.PayloadReference)
	digest := sha256.Sum256([]byte(
		"iam.lifecycle.delivery.v1\x00" + request.PayloadReference,
	))
	requestedAt, timeErr := time.Parse(time.RFC3339Nano, request.RequestedAt)
	messageID, messageErr := provider.smtp.StableMessageID(request.IdempotencyKey)
	if tenantErr != nil || effectErr != nil || referenceErr != nil || timeErr != nil ||
		messageErr != nil || tenantID == uuid.Nil || reference.TenantID() != tenantID ||
		request.PayloadSHA256 != hex.EncodeToString(digest[:]) ||
		request.IdempotencyKey != "iam.lifecycle.delivery:"+request.PayloadReference {
		return lifecycleDeliveryRequest{}, ErrLifecycleDeliveryRuntime
	}
	return lifecycleDeliveryRequest{request: request, tenantID: tenantID,
		requestedAt: requestedAt, messageID: messageID,
		reference: lifecycledeliveryapp.EffectReference{TenantID: tenantID,
			ArtifactRef: request.PayloadReference, EffectID: effectID,
			EffectKey: request.IdempotencyKey, OccurredAt: provider.now().UTC()}}, nil
}
