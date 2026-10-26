package schemaexecutor

import (
	"fmt"
	"slices"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

var supportedVerifyChecks = []schemav1.VerifyCheck{
	schemav1.VerifyConstraints,
	schemav1.VerifyFields,
	schemav1.VerifyFormTable,
	schemav1.VerifyIndexes,
	schemav1.VerifyRelations,
}

func validateVerifyArguments(
	mapping schemav1.IdentifierMapping,
	arguments schemav1.VerifyArguments,
) error {
	if len(arguments.Checks) < 1 || len(arguments.Checks) > len(supportedVerifyChecks) {
		return fmt.Errorf("verify check count is invalid")
	}
	for index, check := range arguments.Checks {
		if !slices.Contains(supportedVerifyChecks, check) {
			return fmt.Errorf("verify check is unknown")
		}
		if index > 0 && arguments.Checks[index-1] >= check {
			return fmt.Errorf("verify checks are not sorted and unique")
		}
	}
	return validateVerifyContracts(mapping, arguments)
}
