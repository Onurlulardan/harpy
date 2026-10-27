package dashboardexport

import "fmt"

type Worker struct {
	store      WorkerStore
	authorizer WorkerAuthorizer
	renderer   Renderer
	artifacts  ArtifactRuntime
	clock      Clock
	ids        IDSource
}

func NewWorker(
	store WorkerStore, authorizer WorkerAuthorizer, renderer Renderer,
	artifacts ArtifactRuntime, clock Clock, ids IDSource,
) (*Worker, error) {
	for _, dependency := range []any{store, authorizer, renderer, artifacts, clock, ids} {
		if nilPort(dependency) {
			return nil, fmt.Errorf("analytics dashboard export worker dependency is unavailable")
		}
	}
	return &Worker{store: store, authorizer: authorizer, renderer: renderer,
		artifacts: artifacts, clock: clock, ids: ids}, nil
}
