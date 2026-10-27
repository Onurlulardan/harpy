package schemaexecutor

import (
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func validateIndexArgumentsV2(
	mapping schemav2.IdentifierMapping,
	arguments schemav2.IndexArguments,
) error {
	if _, found := findIndexIdentifierV2(
		mapping.Indexes, arguments.IndexID, arguments.Generation,
	); !found {
		return fmt.Errorf("index mapping is missing")
	}
	if len(arguments.Fields) < 1 || len(arguments.Fields) > 16 {
		return fmt.Errorf("index field count is invalid")
	}
	seen := map[schemav2.FieldReference]struct{}{}
	for _, field := range arguments.Fields {
		reference := indexReference(field)
		if err := requireFieldV2(mapping, reference); err != nil {
			return err
		}
		if !supportedStorageTypesV2[field.StorageType] {
			return fmt.Errorf("index storage type is invalid")
		}
		if _, exists := seen[reference]; exists {
			return fmt.Errorf("index field is duplicated")
		}
		seen[reference] = struct{}{}
	}
	if arguments.Method != schemav2.IndexBTree && arguments.Method != schemav2.IndexGIN {
		return fmt.Errorf("index method is invalid")
	}
	if arguments.Method == schemav2.IndexGIN &&
		(arguments.Unique || len(arguments.Fields) != 1) {
		return fmt.Errorf("GIN index shape is invalid")
	}
	return validateOperatorClassV2(arguments)
}

func findIndexIdentifierV2(
	values []schemav2.Identifier, logicalID string, generation int64,
) (schemav2.Identifier, bool) {
	for _, value := range values {
		if value.LogicalID == logicalID && value.Generation == generation {
			return value, true
		}
	}
	return schemav2.Identifier{}, false
}

func validateOperatorClassV2(arguments schemav2.IndexArguments) error {
	switch arguments.OperatorClass {
	case schemav2.OperatorClassDefault:
		for _, field := range arguments.Fields {
			if field.Variant != schemav2.VariantValue {
				return fmt.Errorf("default operator class requires value fields")
			}
		}
		return nil
	case schemav2.OperatorClassGINTrigram:
		field := arguments.Fields[0]
		if len(arguments.Fields) != 1 || arguments.Method != schemav2.IndexGIN ||
			arguments.Unique || field.StorageType != schemav2.StorageText ||
			(field.Variant != schemav2.VariantFolded &&
				field.Variant != schemav2.VariantAccent) {
			return fmt.Errorf("trigram operator class shape is invalid")
		}
		return nil
	default:
		return fmt.Errorf("index operator class is invalid")
	}
}

func findIdentifierV2(
	values []schemav2.Identifier,
	logicalID string,
) (schemav2.Identifier, bool) {
	for _, value := range values {
		if value.LogicalID == logicalID {
			return value, true
		}
	}
	return schemav2.Identifier{}, false
}
