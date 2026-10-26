package notifications

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

type deliveryProvider struct {
	identity effectexecution.Identity
	worker   *application.DeliveryWorker
	channel  application.Channel
	now      func() time.Time
}

func newDeliveryProvider(provider string, channel application.Channel,
	worker *application.DeliveryWorker, now func() time.Time,
) (*deliveryProvider, error) {
	identity, err := effectexecution.NewIdentity(provider, "send")
	if err != nil || worker == nil || now == nil ||
		(channel != application.ChannelEmail && channel != application.ChannelTelegram) {
		return nil, application.ErrDeliveryWorker
	}
	return &deliveryProvider{identity: identity, worker: worker,
		channel: channel, now: now}, nil
}

func (provider *deliveryProvider) Identity() effectexecution.Identity {
	if provider == nil {
		return effectexecution.Identity{}
	}
	return provider.identity
}

var _ effectexecution.Provider = (*deliveryProvider)(nil)
