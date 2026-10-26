package application

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type plannedDelivery struct {
	delivery domain.Delivery
	job      *ports.DeliveryJob
	alert    bool
}

type intentPlan struct {
	message    domain.Message
	receipt    domain.EffectReceipt
	deliveries []plannedDelivery
	hint       *ports.RealtimeHint
}

func effectiveChannels(request api.IntentRequest,
	template domain.PublishedTemplate,
) []api.Channel {
	seen := make(map[api.Channel]bool, len(request.Channels)+2)
	for _, channel := range request.Channels {
		seen[channel] = true
	}
	if request.Category != api.CategoryProduct {
		seen[api.ChannelInApp] = true
	}
	if request.Category == api.CategorySecurity || template.MandatoryEmail {
		seen[api.ChannelEmail] = true
	}
	result := make([]api.Channel, 0, len(seen))
	for _, channel := range []api.Channel{api.ChannelInApp, api.ChannelEmail, api.ChannelTelegram} {
		if seen[channel] {
			result = append(result, channel)
		}
	}
	return result
}

func category(value api.Category) domain.Category { return domain.Category(value) }
func channel(value api.Channel) domain.Channel    { return domain.Channel(value) }

func validScope(scope api.Scope) bool {
	if scope == nil || scope.OperationID() == "" {
		return false
	}
	_, _, err := scopeIDs(scope)
	return err == nil
}

func newDeliveryValue(id, messageID domain.ID, selected domain.Channel,
	provider domain.ID, notBefore time.Time, mandatory bool,
) domain.Delivery {
	return domain.Delivery{ID: id, MessageID: messageID, ProviderConfigID: provider,
		Channel: selected, State: domain.DeliveryPending, MaximumAttempts: 8,
		NotBefore: notBefore, Mandatory: mandatory, Revision: 1}
}
