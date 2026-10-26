package authorization

import "bytes"

func validDecisionResults(results []DecisionResult) bool {
	for index, result := range results {
		if result.Validate() != nil {
			return false
		}
		if index > 0 && bytes.Compare(
			results[index-1].requirement[:], result.requirement[:],
		) >= 0 {
			return false
		}
	}
	return true
}

func validDecisionVisibilities(visibilities []Visibility) bool {
	for index, visibility := range visibilities {
		if visibility.Validate() != nil {
			return false
		}
		if index > 0 && bytes.Compare(
			visibilities[index-1].resourceRequirement[:],
			visibility.resourceRequirement[:],
		) >= 0 {
			return false
		}
	}
	return true
}
