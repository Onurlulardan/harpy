package lifecycle

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) resolveDependencies(
	ctx context.Context,
	scope platformtx.TenantScope,
	root packagecontract.Manifest,
) ([]packagecontract.Resolution, error) {
	available := make(map[string][]packagecontract.Manifest)
	registry, err := application.registry.Bind(ctx, scope)
	if err != nil || registry == nil {
		return nil, ErrUnavailable
	}
	queued := append([]packagecontract.Dependency(nil), root.Dependencies...)
	visited := map[string]bool{root.PackageID.String(): true}
	for len(queued) > 0 {
		dependency := queued[0]
		queued = queued[1:]
		key := dependency.PackageID.String()
		if visited[key] {
			continue
		}
		visited[key] = true
		versions, err := registry.CompileCandidates(ctx, key, 64)
		if err != nil {
			if dependency.Kind == packagecontract.DependencyOptional {
				continue
			}
			return nil, ErrDependencyBlocked
		}
		for _, sealed := range versions {
			manifest, err := application.verifySealed(sealed)
			if err != nil {
				return nil, err
			}
			available[key] = append(available[key], manifest)
			queued = append(queued, manifest.Dependencies...)
		}
		if len(visited) > 128 {
			return nil, ErrDependencyBlocked
		}
	}
	resolved, err := packagecontract.ResolveDependencies(root, available)
	if err != nil {
		return nil, ErrDependencyBlocked
	}
	return resolved, nil
}
