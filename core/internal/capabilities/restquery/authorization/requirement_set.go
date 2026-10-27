package authorization

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type resourceRequirement struct {
	requirement Requirement
	catalog     api.FieldCatalog
	path        resolved.RequestPath
}

type requirementSet struct {
	requirements []Requirement
	resources    map[[sha256.Size]byte]resourceRequirement
}

func (set requirementSet) validate() bool {
	if len(set.requirements) < 1 || len(set.requirements) > maximumPolicyRequirements ||
		len(set.resources) < 1 || len(set.resources) > maximumPolicyVisibilities {
		return false
	}
	identities := make(map[[sha256.Size]byte]Requirement, len(set.requirements))
	for _, requirement := range set.requirements {
		if requirement.Validate() != nil {
			return false
		}
		if _, duplicate := identities[requirement.id]; duplicate {
			return false
		}
		identities[requirement.id] = requirement
	}
	resourceCount := 0
	for identity, requirement := range identities {
		if requirement.kind != RequirementResourceRead {
			continue
		}
		resourceCount++
		resource, found := set.resources[identity]
		if !found || !validResourceRequirement(identity, resource) {
			return false
		}
	}
	if resourceCount != len(set.resources) {
		return false
	}
	for identity, resource := range set.resources {
		if _, found := identities[identity]; !found ||
			!validResourceRequirement(identity, resource) {
			return false
		}
	}
	return true
}

func validResourceRequirement(
	identity [sha256.Size]byte,
	resource resourceRequirement,
) bool {
	if resource.requirement.Validate() != nil || resource.catalog.Validate() != nil ||
		resource.path.Validate() != nil || resource.requirement.id != identity ||
		resource.requirement.kind != RequirementResourceRead ||
		resource.requirement.resourcePath != resource.path {
		return false
	}
	reference, err := resource.catalog.Reference()
	return err == nil && reference == resource.requirement.catalog
}
