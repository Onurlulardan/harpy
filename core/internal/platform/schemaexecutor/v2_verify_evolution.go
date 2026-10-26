package schemaexecutor

import (
	"fmt"
	"slices"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func validateEvolutionDataV2(
	mapping schemav2.IdentifierMapping,
	arguments schemav2.VerifyArguments,
) error {
	if !slices.Contains(arguments.Checks, schemav2.VerifyEvolutionData) {
		return nil
	}
	if len(arguments.Evolution) < 1 || len(arguments.Evolution) > 500 {
		return fmt.Errorf("evolution verification count is invalid")
	}
	previous := ""
	for _, contract := range arguments.Evolution {
		key := contract.Target.FieldID + fmt.Sprintf(":%010d", contract.Target.Generation)
		if previous != "" && previous >= key {
			return fmt.Errorf("evolution verifications are not sorted and unique")
		}
		if contract.Source.Variant != schemav2.VariantValue ||
			contract.Target.Variant != schemav2.VariantValue ||
			contract.Source.FieldID != contract.Target.FieldID ||
			contract.Source.Generation >= contract.Target.Generation ||
			contract.Checksum != schemav2.ChecksumXOR64PairV1 ||
			!supportedConvertersV2[contract.Converter] {
			return fmt.Errorf("evolution verification contract is invalid")
		}
		if err := validateTypedLiteralV2(contract.DefaultLiteral); err != nil {
			return err
		}
		if err := requireFieldV2(mapping, contract.Source); err != nil {
			return err
		}
		if err := requireFieldV2(mapping, contract.Target); err != nil {
			return err
		}
		previous = key
	}
	return nil
}
