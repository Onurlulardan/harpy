package adoption

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const operationAcknowledge = "catalog.compatibility.acknowledge"

func (application *Application) AcceptConsumer(
	ctx context.Context, scope platformtx.TenantScope, ack ConsumerAck,
) error {
	if !routePattern.MatchString(ack.RouteKey) || !consumerPattern.MatchString(ack.ConsumerID) ||
		!digestPattern.MatchString(ack.EvidenceSHA256) {
		return ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeCommand,
		operationAcknowledge, "catalog.install")
	if err != nil {
		return err
	}
	now, err := application.now()
	if err != nil {
		return err
	}
	return store.AcceptConsumer(ctx, ack, now)
}
