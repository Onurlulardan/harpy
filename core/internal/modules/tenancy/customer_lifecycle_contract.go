package tenancy

import (
	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
)

type CustomerLifecycleCredential = lifecycle.CredentialEnvelope
type CustomerLifecycleAccessRequest = lifecycle.AccessRequest
type CustomerLifecycleAccessDecision = lifecycle.AccessDecision
type CustomerLifecycleAccessVerifier = lifecycle.AccessVerifier
type CustomerLifecyclePurgeReadiness = lifecycle.PurgeReadiness
type CustomerLifecycleDestructionExecutor = lifecycle.DestructionExecutor
type CustomerLifecycleClock = lifecycle.Clock
type CustomerLifecycleWorkerStore = lifecycle.WorkerStore

type CustomerLifecycleDependencies struct {
	Policy         lifecycle.DeploymentPolicy
	Access         lifecycle.AccessVerifier
	Store          lifecycle.Store
	PurgeReadiness lifecycle.PurgeReadiness
	Clock          lifecycle.Clock
}

func NewCustomerLifecycleApplication(
	dependencies CustomerLifecycleDependencies,
) (*lifecycle.Service, error) {
	return lifecycle.New(lifecycle.Dependencies{
		Policy: dependencies.Policy, Access: dependencies.Access, Store: dependencies.Store,
		PurgeReadiness: dependencies.PurgeReadiness, Clock: dependencies.Clock,
	})
}

func NewCustomerLifecycleWorker(store lifecycle.WorkerStore,
	executor lifecycle.DestructionExecutor, clock lifecycle.Clock,
) (*lifecycle.Worker, error) {
	return lifecycle.NewWorker(lifecycle.WorkerDependencies{
		Store: store, Executor: executor, Clock: clock,
	})
}

var (
	ErrCustomerLifecycleAuthentication = lifecycle.ErrAuthentication
	ErrCustomerLifecycleForbidden      = lifecycle.ErrForbidden
	ErrCustomerLifecycleUnavailable    = lifecycle.ErrUnavailable
)
