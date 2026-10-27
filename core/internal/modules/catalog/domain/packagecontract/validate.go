package packagecontract

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	slugPattern     = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)
	contractPattern = regexp.MustCompile(`^[a-z][a-z0-9.-]{2,63}$`)
	hashPattern     = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

func buildManifest(wire manifestWire, canonical []byte) (Manifest, error) {
	packageID, err := ParsePackageID(wire.PackageID)
	version, versionErr := ParseVersion(wire.Version)
	if err != nil || versionErr != nil || wire.Contract != "catalog-package.v1" ||
		!slugPattern.MatchString(wire.Slug) || !contractPattern.MatchString(wire.PlatformContract) ||
		(wire.Profile != ProfileCore && wire.Profile != ProfileBusiness) ||
		len(wire.Components) == 0 || len(wire.Components) > 256 ||
		len(wire.Dependencies) > 64 || len(wire.Menu) > 512 {
		return Manifest{}, ErrManifestInvalid
	}
	result := Manifest{PackageID: packageID, Slug: wire.Slug, Version: version,
		PlatformContract: wire.PlatformContract, Profile: wire.Profile,
		canonical: append([]byte(nil), canonical...)}
	seenPackages := map[string]bool{packageID.String(): true}
	for _, item := range wire.Dependencies {
		dependency, err := validateDependency(item, seenPackages)
		if err != nil {
			return Manifest{}, err
		}
		result.Dependencies = append(result.Dependencies, dependency)
	}
	componentIDs := make(map[string]ComponentKind, len(wire.Components))
	for _, item := range wire.Components {
		component, err := validateComponent(item, componentIDs)
		if err != nil {
			return Manifest{}, err
		}
		result.Components = append(result.Components, component)
	}
	menuIDs := make(map[string]bool, len(wire.Menu))
	for _, item := range wire.Menu {
		menu, err := validateMenu(item, menuIDs, componentIDs)
		if err != nil {
			return Manifest{}, err
		}
		result.Menu = append(result.Menu, menu)
	}
	if !menuParentsExist(result.Menu, menuIDs) {
		return Manifest{}, ErrManifestInvalid
	}
	return result, nil
}

func validateDependency(wire dependencyWire, seen map[string]bool) (Dependency, error) {
	id, err := ParsePackageID(wire.PackageID)
	rangeValue, rangeErr := ParseRange(wire.Range)
	if err != nil || rangeErr != nil || seen[id.String()] ||
		(wire.Kind != DependencyRequired && wire.Kind != DependencyOptional) {
		return Dependency{}, ErrManifestInvalid
	}
	seen[id.String()] = true
	return Dependency{PackageID: id, Range: rangeValue, Kind: wire.Kind}, nil
}

func validateComponent(wire componentWire, seen map[string]ComponentKind) (Component, error) {
	id, err := ParseComponentID(wire.ID)
	if err != nil || seen[id.String()] != "" || !validComponentOwner(wire.Kind, wire.Owner) ||
		!contractPattern.MatchString(wire.Contract) || !hashPattern.MatchString(wire.ContentSHA256) {
		return Component{}, ErrManifestInvalid
	}
	seen[id.String()] = wire.Kind
	return Component{ID: id, Kind: wire.Kind, Owner: wire.Owner,
		Contract: wire.Contract, ContentHash: wire.ContentSHA256}, nil
}

func validComponentOwner(kind ComponentKind, owner string) bool {
	wanted := map[ComponentKind]string{
		ComponentWorkflow:  "workflow",
		ComponentDashboard: "analytics", ComponentEndpoint: "analytics",
		ComponentIAM: "iam", ComponentDemo: "catalog"}
	return wanted[kind] != "" && wanted[kind] == owner
}

func validExternalURL(raw string) bool {
	value, err := url.Parse(raw)
	return err == nil && value.Scheme == "https" && value.Host != "" &&
		value.User == nil && value.Fragment == "" && !strings.Contains(value.Host, "@")
}
