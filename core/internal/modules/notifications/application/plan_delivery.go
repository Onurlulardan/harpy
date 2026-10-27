package application

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (application *Application) planDelivery(ctx context.Context, stores ports.Stores,
	recipient ports.Recipient, request api.IntentRequest, messageID domain.ID,
	selected api.Channel, now time.Time, mandatoryEmail bool,
	provider ports.ProviderConfig,
) (plannedDelivery, error) {
	preference, found, err := stores.Preferences().LoadPreference(
		ctx, recipient.PrincipalID, category(request.Category), channel(selected),
	)
	if err != nil {
		return plannedDelivery{}, ErrUnavailable
	}
	if !found {
		preference, err = domain.DefaultPreference(
			recipient.PrincipalID, category(request.Category), channel(selected))
	}
	if err != nil {
		return plannedDelivery{}, ErrInvalid
	}
	if selected == api.ChannelEmail {
		preference.VerifiedTarget = recipient.EmailVerified
	} else {
		preference.VerifiedTarget = recipient.TelegramVerified
	}
	decision, err := preference.Decide(now, mandatoryEmail)
	if err != nil {
		return plannedDelivery{}, ErrInvalid
	}
	deliveryID, err := application.nextID(now)
	if err != nil {
		return plannedDelivery{}, err
	}
	base := newDeliveryValue(deliveryID, messageID, channel(selected),
		domain.ID{}, decision.NotBefore, decision.Mandatory)
	if !decision.Enabled {
		value, createErr := domain.NewSuppressedDelivery(base, "PREFERENCE_DISABLED")
		return plannedDelivery{delivery: value}, createErr
	}
	if decision.Mandatory && !preference.VerifiedTarget {
		value, createErr := domain.NewFailedDelivery(base, "RECIPIENT_UNREACHABLE")
		return plannedDelivery{delivery: value, alert: true}, createErr
	}
	if !provider.Enabled || !provider.VersionID.Valid() {
		code := "PROVIDER_UNAVAILABLE"
		alert := decision.Mandatory
		value, createErr := domain.NewFailedDelivery(base, code)
		return plannedDelivery{delivery: value, alert: alert}, createErr
	}
	base.ProviderConfigID = provider.VersionID
	value, err := domain.NewDelivery(base)
	if err != nil {
		return plannedDelivery{}, ErrInvalid
	}
	jobID, err := application.nextID(now)
	if err != nil {
		return plannedDelivery{}, err
	}
	job := &ports.DeliveryJob{ID: jobID, DeliveryID: deliveryID,
		Channel: channel(selected), EffectKey: "notifications.delivery:" + deliveryID.String(),
		RunAt: decision.NotBefore}
	return plannedDelivery{delivery: value, job: job}, nil
}
