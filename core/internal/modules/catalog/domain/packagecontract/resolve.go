package packagecontract

import "sort"

type Resolution struct {
	PackageID PackageID
	Version   Version
	Kind      DependencyKind
}

func ResolveDependencies(root Manifest, available map[string][]Manifest) ([]Resolution, error) {
	if !root.PackageID.Valid() || !root.Version.Valid() {
		return nil, ErrManifestInvalid
	}
	resolved := make(map[string]Resolution)
	visiting := map[string]bool{root.PackageID.String(): true}
	if err := resolveManifest(root, available, resolved, visiting); err != nil {
		return nil, err
	}
	result := make([]Resolution, 0, len(resolved))
	for _, item := range resolved {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].PackageID.String() < result[j].PackageID.String()
	})
	return result, nil
}

func resolveManifest(
	manifest Manifest,
	available map[string][]Manifest,
	resolved map[string]Resolution,
	visiting map[string]bool,
) error {
	for _, dependency := range manifest.Dependencies {
		key := dependency.PackageID.String()
		if visiting[key] {
			return ErrDependencyCycle
		}
		selected, found := selectVersion(dependency, available[key])
		if !found {
			if dependency.Kind == DependencyOptional {
				continue
			}
			return ErrDependencyMissing
		}
		if current, exists := resolved[key]; exists {
			if current.Version.Compare(selected.Version) != 0 {
				return ErrDependencyMissing
			}
			continue
		}
		resolved[key] = Resolution{PackageID: dependency.PackageID,
			Version: selected.Version, Kind: dependency.Kind}
		visiting[key] = true
		if err := resolveManifest(selected, available, resolved, visiting); err != nil {
			return err
		}
		delete(visiting, key)
	}
	return nil
}

func selectVersion(dependency Dependency, candidates []Manifest) (Manifest, bool) {
	var selected Manifest
	found := false
	for _, candidate := range candidates {
		if candidate.PackageID.String() != dependency.PackageID.String() ||
			!dependency.Range.Contains(candidate.Version) {
			continue
		}
		if !found || candidate.Version.Compare(selected.Version) > 0 {
			selected, found = candidate, true
		}
	}
	return selected, found
}
