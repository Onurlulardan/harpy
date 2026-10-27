package customerprovisioning

import "fmt"

type Dependencies struct {
	Policy    DeploymentPolicy
	Access    AccessVerifier
	Offering  OfferingResolver
	Authority AuthorityPlanner
	Setup     SetupPreparer
	Store     OperationStore
	Clock     Clock
}

type Coordinator struct{ dependencies Dependencies }

func New(dependencies Dependencies) (*Coordinator, error) {
	if dependencies.Policy == nil || dependencies.Access == nil ||
		dependencies.Offering == nil || dependencies.Authority == nil ||
		dependencies.Setup == nil || dependencies.Store == nil || dependencies.Clock == nil {
		return nil, fmt.Errorf("customer provisioning dependencies are incomplete")
	}
	return &Coordinator{dependencies: dependencies}, nil
}
