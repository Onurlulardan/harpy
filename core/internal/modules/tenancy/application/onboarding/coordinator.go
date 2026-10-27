package onboarding

import tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"

type Coordinator struct {
	store OperationStore
	unit  ActivationUnitOfWork
}

func New(store OperationStore, unit ActivationUnitOfWork) (*Coordinator, error) {
	if store == nil || unit == nil {
		return nil, tenantapi.ErrOnboardingUnavailable
	}
	return &Coordinator{store: store, unit: unit}, nil
}

func (coordinator *Coordinator) usable() bool {
	return coordinator != nil && coordinator.store != nil && coordinator.unit != nil
}
