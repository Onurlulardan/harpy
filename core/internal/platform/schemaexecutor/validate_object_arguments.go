package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func validateObjectArguments(
	operation schemav1.OperationID,
	mapping schemav1.IdentifierMapping,
	objectID string,
) error {
	var found bool
	switch operation {
	case schemav1.DropRelationTable:
		_, found = findIdentifier(mapping.Relations, objectID)
	case schemav1.CleanupInvalidIndex, schemav1.DropIndex, schemav1.DropIndexConcurrent:
		_, found = findIdentifier(mapping.Indexes, objectID)
	case schemav1.DropConstraint, schemav1.ValidateConstraint:
		_, found = findIdentifier(mapping.Constraints, objectID)
	default:
		return fmt.Errorf("object operation is not registered")
	}
	if !found {
		return fmt.Errorf("operation object mapping is missing")
	}
	return nil
}
