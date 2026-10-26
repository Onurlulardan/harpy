package queryhttp

import queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"

type dependencyView struct {
	OwnerDomain  string `json:"owner_domain"`
	ResourceKind string `json:"resource_kind"`
	ResourceID   string `json:"resource_id"`
	VersionID    string `json:"version_id"`
}

type publishedView struct {
	view
	Dependencies []dependencyView `json:"dependencies"`
}

func presentPublished(value queryapp.PublishedSnapshot) publishedView {
	dependencies := make([]dependencyView, len(value.Dependencies))
	for index, dependency := range value.Dependencies {
		dependencies[index] = dependencyView{
			OwnerDomain:  dependency.OwnerDomain,
			ResourceKind: dependency.ResourceKind,
			ResourceID:   dependency.ResourceID.String(),
			VersionID:    dependency.VersionID.String(),
		}
	}
	return publishedView{
		view:         present(value.Snapshot),
		Dependencies: dependencies,
	}
}
