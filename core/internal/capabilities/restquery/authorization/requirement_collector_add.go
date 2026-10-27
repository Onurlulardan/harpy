package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func (collector *requirementCollector) add(
	kind RequirementKind,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	requestPath resolved.RequestPath,
	fieldID uuid.UUID,
	relationID uuid.UUID,
) (Requirement, error) {
	if collector == nil || catalog.Validate() != nil ||
		len(collector.requirements) >= maximumPolicyRequirements {
		return Requirement{}, errPolicyEvaluationContract
	}
	reference, err := catalog.Reference()
	if err != nil {
		return Requirement{}, errPolicyEvaluationContract
	}
	requirement, err := newRequirement(
		kind, reference, resourcePath, requestPath, fieldID, relationID,
	)
	if err != nil {
		return Requirement{}, errPolicyEvaluationContract
	}
	if _, duplicate := collector.identities[requirement.id]; duplicate {
		return Requirement{}, errPolicyEvaluationContract
	}
	collector.identities[requirement.id] = struct{}{}
	collector.requirements = append(collector.requirements, requirement)
	return requirement, nil
}

func (collector *requirementCollector) build() (requirementSet, error) {
	if collector == nil {
		return requirementSet{}, errPolicyEvaluationContract
	}
	resources := make(map[[32]byte]resourceRequirement, len(collector.resources))
	for identity, resource := range collector.resources {
		resources[identity] = resource
	}
	set := requirementSet{
		requirements: append([]Requirement{}, collector.requirements...),
		resources:    resources,
	}
	if !set.validate() {
		return requirementSet{}, errPolicyEvaluationContract
	}
	return set, nil
}
