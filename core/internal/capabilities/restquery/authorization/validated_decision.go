package authorization

import "crypto/sha256"

type validatedDecision struct {
	requestBinding  [sha256.Size]byte
	contextRevision [sha256.Size]byte
	denied          []Requirement
	visibilities    []Visibility
}

func validatePolicyDecision(
	request PolicyRequest,
	decision PolicyDecision,
) (validatedDecision, error) {
	if request.Validate() != nil {
		return validatedDecision{}, errPolicyDecisionContract
	}
	return validateDecision(request.binding, request.requirements, decision)
}

func validateDecision(
	binding [sha256.Size]byte,
	requirements []Requirement,
	decision PolicyDecision,
) (validatedDecision, error) {
	if binding == [sha256.Size]byte{} || decision.Validate() != nil ||
		!sameDigest(binding, decision.requestBinding) ||
		len(requirements) != len(decision.results) {
		return validatedDecision{}, errPolicyDecisionContract
	}
	denied := make([]Requirement, 0)
	allowedResources := make(map[[sha256.Size]byte]struct{})
	for index, requirement := range requirements {
		if requirement.Validate() != nil {
			return validatedDecision{}, errPolicyDecisionContract
		}
		result := decision.results[index]
		if !sameDigest(requirement.id, result.requirement) {
			return validatedDecision{}, errPolicyDecisionContract
		}
		if !result.allowed {
			denied = append(denied, requirement)
			continue
		}
		if requirement.kind == RequirementResourceRead {
			allowedResources[requirement.id] = struct{}{}
		}
	}
	if len(allowedResources) != len(decision.visibilities) {
		return validatedDecision{}, errPolicyDecisionContract
	}
	for _, visibility := range decision.visibilities {
		if _, found := allowedResources[visibility.resourceRequirement]; !found {
			return validatedDecision{}, errPolicyDecisionContract
		}
	}
	return validatedDecision{
		requestBinding:  binding,
		contextRevision: decision.contextRevision,
		denied:          append([]Requirement{}, denied...),
		visibilities:    cloneVisibilities(decision.visibilities),
	}, nil
}

func cloneVisibilities(values []Visibility) []Visibility {
	result := make([]Visibility, len(values))
	for index := range values {
		result[index] = values[index].clone()
	}
	return result
}
