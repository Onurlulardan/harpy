package iam

import "fmt"

type CatalogAccessRuntime struct{ shared *accessRequestDependencies }

func NewCatalogAccessRuntime(
	shared *accessRequestDependencies,
) (*CatalogAccessRuntime, error) {
	if shared == nil || shared.pool == nil || shared.coordinator == nil ||
		shared.security == nil || shared.clock == nil {
		return nil, fmt.Errorf("IAM Catalog access dependencies are required")
	}
	return &CatalogAccessRuntime{shared: shared}, nil
}
