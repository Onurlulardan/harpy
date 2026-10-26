package menu

import "fmt"

type Application struct {
	binder     StoreBinder
	authorizer Authorizer
	targets    TargetAuthorizer
	clock      Clock
	ids        IDSource
}

func NewApplication(
	binder StoreBinder, authorizer Authorizer, targets TargetAuthorizer,
	clock Clock, ids IDSource,
) (*Application, error) {
	if binder == nil || authorizer == nil || targets == nil || clock == nil || ids == nil {
		return nil, fmt.Errorf("catalog menu dependencies are required")
	}
	return &Application{binder: binder, authorizer: authorizer,
		targets: targets, clock: clock, ids: ids}, nil
}
