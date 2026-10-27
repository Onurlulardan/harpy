package authorization

import "crypto/sha256"

func (evaluation policyEvaluation) validate() bool {
	if evaluation.request.Validate() != nil ||
		evaluation.owner != evaluation.request.owner ||
		!sameDigest(evaluation.decision.requestBinding, evaluation.request.binding) ||
		evaluation.decision.contextRevision == [sha256.Size]byte{} {
		return false
	}
	requestRequirements := make(map[[sha256.Size]byte]Requirement)
	for _, requirement := range evaluation.request.requirements {
		requestRequirements[requirement.id] = requirement
	}
	denied := make(map[[sha256.Size]byte]struct{})
	for _, requirement := range evaluation.decision.denied {
		if requirement.Validate() != nil {
			return false
		}
		if _, found := requestRequirements[requirement.id]; !found {
			return false
		}
		if _, duplicate := denied[requirement.id]; duplicate {
			return false
		}
		denied[requirement.id] = struct{}{}
	}
	expected := make(map[[sha256.Size]byte]struct{})
	for _, requirement := range evaluation.request.requirements {
		if requirement.kind != RequirementResourceRead {
			continue
		}
		if _, isDenied := denied[requirement.id]; !isDenied {
			expected[requirement.id] = struct{}{}
		}
	}
	if len(expected) != len(evaluation.decision.visibilities) {
		return false
	}
	for _, visibility := range evaluation.decision.visibilities {
		if visibility.Validate() != nil {
			return false
		}
		if _, found := expected[visibility.resourceRequirement]; !found {
			return false
		}
		delete(expected, visibility.resourceRequirement)
	}
	return len(expected) == 0
}
