package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func validateIndexArguments(
	mapping schemav1.IdentifierMapping,
	arguments schemav1.IndexArguments,
) error {
	if _, found := findIdentifier(mapping.Indexes, arguments.IndexID); !found {
		return fmt.Errorf("index mapping is missing")
	}
	if len(arguments.FieldIDs) < 1 || len(arguments.FieldIDs) > 16 {
		return fmt.Errorf("index field count is invalid")
	}
	seen := map[string]struct{}{}
	for _, fieldID := range arguments.FieldIDs {
		if err := requireField(mapping, fieldID); err != nil {
			return err
		}
		if _, exists := seen[fieldID]; exists {
			return fmt.Errorf("index field is duplicated")
		}
		seen[fieldID] = struct{}{}
	}
	if arguments.Method != schemav1.IndexBTree && arguments.Method != schemav1.IndexGIN {
		return fmt.Errorf("index method is invalid")
	}
	if arguments.Method == schemav1.IndexGIN &&
		(arguments.Unique || len(arguments.FieldIDs) != 1) {
		return fmt.Errorf("GIN index shape is invalid")
	}
	return nil
}
