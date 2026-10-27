package deadletterdelivery

import "github.com/Onurlulardan/harpy/core/internal/platform/admission"

func New(
	connector Connector,
	settings Settings,
	gate *admission.Gate,
) (*Runner, error) {
	return newWithJitter(connector, settings, runtimeJitter{}, gate)
}

func newWithJitter(
	connector Connector,
	settings Settings,
	jitter jitterSource,
	gate *admission.Gate,
) (*Runner, error) {
	if connector == nil || jitter == nil || gate == nil || !settings.valid() {
		return nil, ErrDependencies
	}
	return &Runner{
		connector: connector,
		settings:  settings,
		jitter:    jitter,
		admission: gate,
		lifecycle: newRunnerLifecycle(),
	}, nil
}
