package schemaexecutor

import (
	"fmt"
	"slices"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func validatePlanContractV2(plan schemav2.Plan) error {
	if plan.ContractVersion != schemav2.PlanContractVersion || plan.Owner != "forms" {
		return fmt.Errorf("plan contract or owner differs")
	}
	if plan.Compiler.Version != schemav2.CompilerVersion ||
		plan.Compiler.PostgreSQLMajor != schemav2.RequiredPostgreSQLMajor {
		return fmt.Errorf("compiler is not allowlisted")
	}
	if _, err := canonicalUUID(plan.TenantID); err != nil {
		return fmt.Errorf("tenant identity is invalid")
	}
	if _, err := canonicalUUID(plan.Resource.ID); err != nil ||
		plan.Resource.Revision < 1 || plan.Resource.Revision > maximumSafeJSONInteger {
		return fmt.Errorf("resource identity is invalid")
	}
	if plan.Resource.Kind != schemav2.ResourceFormStorage &&
		plan.Resource.Kind != schemav2.ResourceFormPurge {
		return fmt.Errorf("resource kind is invalid")
	}
	if len(plan.Steps) < 1 || len(plan.Steps) > schemav2.MaximumPlanSteps {
		return fmt.Errorf("plan step budget differs")
	}
	return validateOperationAllowlistV2(plan)
}

func validateOperationAllowlistV2(plan schemav2.Plan) error {
	if len(plan.AllowedOperations) < 1 ||
		len(plan.AllowedOperations) > len(v2OperationContracts) {
		return fmt.Errorf("operation allowlist size is invalid")
	}
	seen := make(map[schemav2.OperationID]bool, len(plan.AllowedOperations))
	for index, operation := range plan.AllowedOperations {
		if _, supported := v2OperationContracts[operation]; !supported {
			return fmt.Errorf("operation allowlist contains unknown operation")
		}
		if index > 0 && plan.AllowedOperations[index-1] >= operation {
			return fmt.Errorf("operation allowlist is not sorted and unique")
		}
		seen[operation] = true
	}
	for _, step := range plan.Steps {
		if !seen[step.OperationID] ||
			!slices.Contains(plan.AllowedOperations, step.OperationID) {
			return fmt.Errorf("step operation is outside plan allowlist")
		}
	}
	return nil
}

func validateResourceOperationV2(
	resource schemav2.ResourceKind,
	operation schemav2.OperationID,
) error {
	if operation == schemav2.DropFormTable && resource != schemav2.ResourceFormPurge {
		return fmt.Errorf("form table drop requires form-purge resource")
	}
	return nil
}
