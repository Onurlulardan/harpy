package notifications

import (
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

const (
	emailDeliveryJobKind    = "notifications.email.delivery"
	telegramDeliveryJobKind = "notifications.telegram.delivery"
)

func NewDeliveryJobProcessors(release platformasync.ReleaseCatalogs,
	store asyncpostgres.WorkerStore, worker *application.DeliveryWorker,
) ([]jobexecution.Processor, error) {
	if store == nil || worker == nil {
		return nil, application.ErrDeliveryWorker
	}
	jobs, _, err := release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read notification delivery catalog: %w", err)
	}
	smtp, err := newDeliveryProvider("notifications.smtp", application.ChannelEmail,
		worker, time.Now)
	if err != nil {
		return nil, err
	}
	telegram, err := newDeliveryProvider("notifications.telegram",
		application.ChannelTelegram, worker, time.Now)
	if err != nil {
		return nil, err
	}
	registry, err := effectexecution.NewRegistry(
		[]effectexecution.Provider{smtp, telegram})
	if err != nil {
		return nil, err
	}
	preEffect, err := asyncpostgres.BindExternalPreEffect(store)
	if err != nil {
		return nil, err
	}
	definitions := []struct{ kind, provider string }{
		{emailDeliveryJobKind, "notifications.smtp"},
		{telegramDeliveryJobKind, "notifications.telegram"},
	}
	processors := make([]jobexecution.Processor, 0, len(definitions))
	for _, definition := range definitions {
		policy, resolveErr := jobs.Resolve(definition.kind, 1)
		if resolveErr != nil || !validDeliveryPolicy(policy) {
			return nil, fmt.Errorf("notification delivery policy is invalid")
		}
		processor, buildErr := jobexecution.NewExternalProcessor(policy,
			deliveryPlanner{kind: definition.kind, provider: definition.provider},
			deliveryPreEffect{store: preEffect}, store, registry)
		if buildErr != nil {
			return nil, buildErr
		}
		processors = append(processors, processor)
	}
	return processors, nil
}

func validDeliveryPolicy(policy jobsv1.KindPolicy) bool {
	return policy.OwnerDomain == "notifications" &&
		policy.EffectClass == jobsv1.EffectExternal && policy.MaximumAttempts == 8
}
