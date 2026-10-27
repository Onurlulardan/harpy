package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (value *testStores) InsertDelivery(_ context.Context, delivery domain.Delivery) error {
	value.deliveries[delivery.ID] = delivery
	return nil
}

func (*testStores) InsertAttempt(context.Context, ports.DeliveryAttempt) error { return nil }
func (*testStores) InsertReceipt(context.Context, ports.DeliveryReceipt) error { return nil }

func (value *testStores) AppendHint(_ context.Context, hint ports.RealtimeHint) error {
	value.hints = append(value.hints, hint)
	return nil
}

func (value *testStores) AppendAudit(_ context.Context, action string,
	_, _ domain.ID, _ int64,
) error {
	value.audits = append(value.audits, action)
	return nil
}

func (value *testStores) EnqueueDelivery(_ context.Context, job ports.DeliveryJob) error {
	value.jobs = append(value.jobs, job)
	return nil
}

func (value *testStores) AppendOperatorAlert(_ context.Context, code string, _ domain.ID) error {
	value.alerts = append(value.alerts, code)
	return nil
}
