package query

import (
	analyticsdependency "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dependency"
)

func dependencyManifest(
	requested []DependencyRequest, resolved []DependencySnapshot, revisionBound bool,
) (analyticsdependency.Manifest, error) {
	if len(requested) == 0 || len(requested) != len(resolved) {
		return analyticsdependency.Manifest{}, ErrOwnerNotV2
	}
	entries := make([]analyticsdependency.Entry, 0, len(resolved))
	seen := map[DependencyRequest]struct{}{}
	for _, value := range resolved {
		if !value.WriterV2 || value.Retired || value.OwnerGeneration == 0 {
			return analyticsdependency.Manifest{}, ErrOwnerNotV2
		}
		if _, exists := seen[value.DependencyRequest]; exists {
			return analyticsdependency.Manifest{}, ErrInvalid
		}
		seen[value.DependencyRequest] = struct{}{}
		entry, err := analyticsdependency.NewEntry(value.OwnerDomain, value.ResourceKind,
			[16]byte(value.ResourceID), [16]byte(value.VersionID), value.CatalogVersion,
			value.QueryContractVersion, value.PolicyVersion,
			value.DataRevisionCapability, value.DependencySHA256)
		if err != nil {
			return analyticsdependency.Manifest{}, ErrInvalid
		}
		entries = append(entries, entry)
	}
	for _, request := range requested {
		if _, exists := seen[request]; !exists {
			return analyticsdependency.Manifest{}, ErrOwnerNotV2
		}
	}
	manifest, err := analyticsdependency.NewManifest(entries, revisionBound)
	if err != nil {
		return analyticsdependency.Manifest{}, ErrInvalid
	}
	return manifest, nil
}
