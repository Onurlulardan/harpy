package schemaexecutor

import (
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

var supportedConvertersV2 = map[schemav2.Converter]bool{
	schemav2.ConvertIdentity: true, schemav2.ConvertIntegerDecimal: true,
	schemav2.ConvertDecimalWiden: true,
	schemav2.ConvertIntegerText:  true, schemav2.ConvertDecimalText: true,
	schemav2.ConvertDateTimestamp: true, schemav2.ConvertUUIDText: true,
}

func validateBackfillArgumentsV2(
	mapping schemav2.IdentifierMapping,
	arguments schemav2.BackfillArguments,
) error {
	if arguments.Source == arguments.Target && arguments.DefaultLiteral == nil {
		return fmt.Errorf("backfill source and target identities must differ")
	}
	if arguments.Source.Variant != schemav2.VariantValue ||
		arguments.Target.Variant != schemav2.VariantValue {
		return fmt.Errorf("typed converter backfill requires value variants")
	}
	if err := requireFieldV2(mapping, arguments.Source); err != nil {
		return err
	}
	if err := requireFieldV2(mapping, arguments.Target); err != nil {
		return err
	}
	if !supportedConvertersV2[arguments.Converter] {
		return fmt.Errorf("backfill converter is invalid")
	}
	return validateTypedLiteralV2(arguments.DefaultLiteral)
}

func validateGenerationArgumentsV2(arguments schemav2.GenerationArguments) error {
	if _, err := canonicalUUID(arguments.FormVersionID); err != nil {
		return fmt.Errorf("form version ID is invalid")
	}
	if arguments.SourceGeneration < 1 || arguments.TargetGeneration < 1 ||
		arguments.SourceGeneration > schemav2.MaximumFieldGeneration ||
		arguments.TargetGeneration > schemav2.MaximumFieldGeneration ||
		arguments.SourceGeneration == arguments.TargetGeneration {
		return fmt.Errorf("generation handoff is invalid")
	}
	return nil
}

func validateRetentionArgumentsV2(arguments schemav2.RetentionArguments) error {
	if _, err := canonicalUUID(arguments.FormVersionID); err != nil {
		return fmt.Errorf("form version ID is invalid")
	}
	if arguments.RetainedGeneration < 1 ||
		arguments.RetainedGeneration > schemav2.MaximumFieldGeneration {
		return fmt.Errorf("retained generation is invalid")
	}
	if arguments.MinimumRetentionDays != 30 || arguments.ZeroReferenceDays != 7 {
		return fmt.Errorf("retention policy differs from the accepted contract")
	}
	return nil
}
