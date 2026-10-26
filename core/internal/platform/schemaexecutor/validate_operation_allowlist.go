package schemaexecutor

import (
	"fmt"
	"slices"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func validateOperationAllowlist(plan schemav1.Plan) error {
	if len(plan.AllowedOperations) < 1 ||
		len(plan.AllowedOperations) > len(operationContracts) {
		return fmt.Errorf("operation allowlist size is invalid")
	}
	seen := make(map[schemav1.OperationID]bool, len(plan.AllowedOperations))
	for index, operation := range plan.AllowedOperations {
		if _, supported := operationContracts[operation]; !supported {
			return fmt.Errorf("operation allowlist contains unknown operation")
		}
		if index > 0 && plan.AllowedOperations[index-1] >= operation {
			return fmt.Errorf("operation allowlist is not sorted and unique")
		}
		seen[operation] = true
	}
	for _, step := range plan.Steps {
		if !seen[step.OperationID] || !slices.Contains(plan.AllowedOperations, step.OperationID) {
			return fmt.Errorf("step operation is outside plan allowlist")
		}
	}
	return nil
}
