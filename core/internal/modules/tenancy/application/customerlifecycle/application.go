package customerlifecycle

import (
	"fmt"
	"reflect"
)

type Dependencies struct {
	Policy         DeploymentPolicy
	Access         AccessVerifier
	Store          Store
	PurgeReadiness PurgeReadiness
	Clock          Clock
}

type Service struct{ dependencies Dependencies }

func New(dependencies Dependencies) (*Service, error) {
	if nilDependency(dependencies.Policy) || nilDependency(dependencies.Access) ||
		nilDependency(dependencies.Store) || nilDependency(dependencies.PurgeReadiness) ||
		nilDependency(dependencies.Clock) {
		return nil, fmt.Errorf("customer lifecycle dependencies are incomplete")
	}
	return &Service{dependencies: dependencies}, nil
}

func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}

type WorkerDependencies struct {
	Store    WorkerStore
	Executor DestructionExecutor
	Clock    Clock
}

type Worker struct{ dependencies WorkerDependencies }

func NewWorker(dependencies WorkerDependencies) (*Worker, error) {
	if nilDependency(dependencies.Store) || nilDependency(dependencies.Executor) ||
		nilDependency(dependencies.Clock) {
		return nil, fmt.Errorf("customer lifecycle worker dependencies are incomplete")
	}
	return &Worker{dependencies: dependencies}, nil
}
