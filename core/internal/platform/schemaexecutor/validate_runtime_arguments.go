package schemaexecutor

import (
	"fmt"
	"time"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func validateGenerationArguments(arguments schemav1.GenerationArguments) error {
	if _, err := canonicalUUID(arguments.FormVersionID); err != nil {
		return fmt.Errorf("form version ID is invalid")
	}
	if arguments.SourceGeneration < 1 || arguments.TargetGeneration < 1 ||
		arguments.SourceGeneration > maximumFieldGeneration ||
		arguments.TargetGeneration > maximumFieldGeneration ||
		arguments.SourceGeneration == arguments.TargetGeneration {
		return fmt.Errorf("generation handoff is invalid")
	}
	return nil
}

func validateRetentionArguments(arguments schemav1.RetentionArguments) error {
	if _, err := canonicalUUID(arguments.FormVersionID); err != nil {
		return fmt.Errorf("form version ID is invalid")
	}
	if arguments.RetainedGeneration < 1 ||
		arguments.RetainedGeneration > maximumFieldGeneration {
		return fmt.Errorf("retained generation is invalid")
	}
	parsed, err := time.Parse(time.RFC3339, arguments.ContractNotBefore)
	if err != nil || parsed.Format(time.RFC3339) != arguments.ContractNotBefore ||
		parsed.Location() != time.UTC {
		return fmt.Errorf("contract-not-before timestamp is not canonical UTC")
	}
	return nil
}
