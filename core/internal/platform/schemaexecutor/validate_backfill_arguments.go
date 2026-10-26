package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

var supportedConverters = map[schemav1.Converter]bool{
	schemav1.ConvertIdentity: true, schemav1.ConvertIntegerDecimal: true,
	schemav1.ConvertIntegerText: true, schemav1.ConvertDecimalText: true,
	schemav1.ConvertDateTimestamp: true, schemav1.ConvertUUIDText: true,
}

func validateBackfillArguments(
	mapping schemav1.IdentifierMapping,
	arguments schemav1.BackfillArguments,
) error {
	if arguments.SourceFieldID == arguments.TargetFieldID &&
		arguments.SourceGeneration == arguments.TargetGeneration {
		return fmt.Errorf("backfill source and target identities must differ")
	}
	if err := requireFieldGeneration(
		mapping, arguments.SourceFieldID, arguments.SourceGeneration,
	); err != nil {
		return err
	}
	if err := requireFieldGeneration(
		mapping, arguments.TargetFieldID, arguments.TargetGeneration,
	); err != nil {
		return err
	}
	if !supportedConverters[arguments.Converter] {
		return fmt.Errorf("backfill converter is invalid")
	}
	return nil
}

func requireFieldGeneration(
	mapping schemav1.IdentifierMapping,
	fieldID string,
	generation int64,
) error {
	if _, err := canonicalUUID(fieldID); err != nil {
		return fmt.Errorf("field ID is invalid")
	}
	if generation < 1 || generation > maximumFieldGeneration {
		return fmt.Errorf("field generation is invalid")
	}
	if _, found := findFieldGeneration(mapping.Fields, fieldID, generation); !found {
		return fmt.Errorf("field generation mapping is missing")
	}
	return nil
}
