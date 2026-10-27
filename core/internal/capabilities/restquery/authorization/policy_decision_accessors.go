package authorization

import "crypto/sha256"

func (decision PolicyDecision) RequestBindingSHA256() ([sha256.Size]byte, error) {
	if err := decision.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	return decision.requestBinding, nil
}

func (decision PolicyDecision) ContextRevisionSHA256() ([sha256.Size]byte, error) {
	if err := decision.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	return decision.contextRevision, nil
}

func (decision PolicyDecision) Results() ([]DecisionResult, error) {
	if err := decision.Validate(); err != nil {
		return nil, err
	}
	return append([]DecisionResult{}, decision.results...), nil
}

func (decision PolicyDecision) Visibilities() ([]Visibility, error) {
	if err := decision.Validate(); err != nil {
		return nil, err
	}
	result := make([]Visibility, len(decision.visibilities))
	for index := range decision.visibilities {
		result[index] = decision.visibilities[index].clone()
	}
	return result, nil
}
