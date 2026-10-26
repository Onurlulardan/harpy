package notifications

import (
	"fmt"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
)

func (provider *deliveryProvider) validateRequest(
	request effectsv1.Request,
) (deliveryReference, error) {
	if provider == nil || request.Validate() != nil || request.TenantID == nil ||
		request.Authority.Domain != "notifications" || request.Operation != "send" ||
		request.Provider != provider.identity.Provider() || request.EffectRevision != 1 {
		return deliveryReference{}, fmt.Errorf("notification delivery request is invalid")
	}
	kind := emailDeliveryJobKind
	if provider.channel == application.ChannelTelegram {
		kind = telegramDeliveryJobKind
	}
	reference, err := parseDeliveryReference(request.PayloadReference)
	if err != nil || request.EffectKind != kind ||
		request.PayloadSHA256 != deliveryReferenceDigest(request.PayloadReference) ||
		*request.TenantID != reference.tenantID.String() ||
		request.IdempotencyKey != "notifications.delivery:"+reference.deliveryID.String() {
		return deliveryReference{}, fmt.Errorf("notification delivery request is invalid")
	}
	return reference, nil
}
