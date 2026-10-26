package outboxdelivery

import (
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

func New(
	store asyncpostgres.PublisherStore,
	catalog eventsv1.ProducerCatalog,
	connector Connector,
	settings Settings,
	gate *admission.Gate,
) (*Runner, error) {
	return newWithJitter(store, catalog, connector, settings, gate, runtimeJitter{})
}

func newWithJitter(
	store asyncpostgres.PublisherStore,
	catalog eventsv1.ProducerCatalog,
	connector Connector,
	settings Settings,
	gate *admission.Gate,
	jitter jitterSource,
) (*Runner, error) {
	if store == nil || connector == nil || gate == nil || jitter == nil || !settings.valid() {
		return nil, ErrDependencies
	}
	validated, err := eventsv1.NewProducerCatalog(catalog.Policies())
	if err != nil {
		return nil, fmt.Errorf("validate outbox producer catalog: %w", err)
	}
	id := uuid.New()
	return &Runner{
		store: store, catalog: validated, connector: connector,
		settings: settings, admission: gate, publisherID: id, jitter: jitter,
		lifecycle: newRunnerLifecycle(),
	}, nil
}
