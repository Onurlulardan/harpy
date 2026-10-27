package application

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (worker *DeliveryWorker) execute(ctx context.Context, tenantID domain.ID,
	work ports.DeliveryWork, startedAt time.Time,
) (DeliveryProcessResult, error) {
	if !work.Provider.Enabled {
		return worker.finalize(ctx, tenantID, work, startedAt,
			ports.ProviderResult{Outcome: ports.ProviderPermanent,
				Code: "PROVIDER_DISABLED", ResponseClass: "provider_policy",
				CompletedAt: worker.clock.Now().UTC()})
	}
	target, err := worker.targets.ResolveDeliveryTarget(ctx, tenantID,
		work.Message.RecipientID, work.Delivery.Channel)
	if err != nil || len(target) == 0 {
		return worker.finalize(ctx, tenantID, work, startedAt,
			ports.ProviderResult{Outcome: ports.ProviderPermanent,
				Code: "RECIPIENT_UNREACHABLE", ResponseClass: "recipient",
				CompletedAt: worker.clock.Now().UTC()})
	}
	defer clear(target)
	credential, err := worker.opener.Open(ctx, work.Secret)
	if err != nil || len(credential) == 0 {
		return worker.finalize(ctx, tenantID, work, startedAt,
			ports.ProviderResult{Outcome: ports.ProviderTransient,
				Code: "SECRET_UNAVAILABLE", ResponseClass: "pre_effect",
				CompletedAt: worker.clock.Now().UTC()})
	}
	defer clear(credential)
	sender, err := worker.registry.Resolve(work.Provider.Kind)
	if err != nil || sender == nil {
		return worker.finalize(ctx, tenantID, work, startedAt,
			ports.ProviderResult{Outcome: ports.ProviderTransient,
				Code: "PROVIDER_UNAVAILABLE", ResponseClass: "pre_effect",
				CompletedAt: worker.clock.Now().UTC()})
	}
	result := sender.Send(ctx, ports.ProviderRequest{DeliveryID: work.Delivery.ID,
		Target: target, Subject: work.Message.Subject, Body: work.Message.Body,
		TemplateKind: work.Message.TemplateKind,
		Metadata:     work.Provider.Metadata, Credential: credential, Requested: startedAt})
	return worker.finalize(ctx, tenantID, work, startedAt, result)
}
