package eventdelivery

import (
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

// New builds the production RabbitMQ event workload. An empty registry is an
// explicit no-op and does not require a usable broker credential.
func New(
	settings config.RabbitMQ,
	producers eventsv1.ProducerCatalog,
	registry Registry,
	gate *admission.Gate,
) (*Runner, error) {
	validatedRegistry, err := cloneRegistry(registry)
	if err != nil {
		return nil, fmt.Errorf("validate event handler registry: %w", err)
	}
	validatedProducers, err := eventsv1.NewProducerCatalog(producers.Policies())
	if err != nil {
		return nil, fmt.Errorf("validate event producer catalog: %w", err)
	}
	if err := validateProducerCoverage(validatedProducers, validatedRegistry.Catalog()); err != nil {
		return nil, err
	}
	if validatedRegistry.Empty() {
		return newRunner(validatedRegistry, nil, gate, runtimeJitter{}, waitContext)
	}
	if settings.URL().Value() == "" {
		return nil, ErrRunnerDependencies
	}
	connector := &rabbitConnector{
		settings: settings, producers: validatedProducers,
		consumers: validatedRegistry.Catalog(),
	}
	return newRunner(validatedRegistry, connector, gate, runtimeJitter{}, waitContext)
}

func newRunner(
	registry Registry,
	connector sessionConnector,
	gate *admission.Gate,
	jitter jitterSource,
	wait waitFunction,
) (*Runner, error) {
	validated, err := cloneRegistry(registry)
	if err != nil || gate == nil || jitter == nil || wait == nil ||
		(!validated.Empty() && connector == nil) {
		return nil, ErrRunnerDependencies
	}
	return &Runner{
		registry: validated, connector: connector, admission: gate, jitter: jitter, wait: wait,
		lifecycle: newRunnerLifecycle(),
	}, nil
}

func validateProducerCoverage(
	producers eventsv1.ProducerCatalog,
	consumers eventsv1.ConsumerCatalog,
) error {
	for _, policy := range consumers.Policies() {
		if _, err := producers.Resolve(policy.EventType, policy.EventVersion); err != nil {
			return fmt.Errorf(
				"consumer %s version %d event %s version %d: producer contract is missing",
				policy.Consumer, policy.ConsumerVersion, policy.EventType, policy.EventVersion,
			)
		}
	}
	return nil
}
