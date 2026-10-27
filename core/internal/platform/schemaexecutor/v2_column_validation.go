package schemaexecutor

import (
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

var supportedStorageTypesV2 = map[schemav2.StorageType]bool{
	schemav2.StorageText: true, schemav2.StorageInteger: true,
	schemav2.StorageDecimal: true, schemav2.StorageBoolean: true,
	schemav2.StorageDate: true, schemav2.StorageTimestamp: true,
	schemav2.StorageUUID: true, schemav2.StorageEnum: true,
	schemav2.StorageRelation: true, schemav2.StorageTextList: true,
	schemav2.StorageUUIDList: true, schemav2.StorageJSON: true,
}

func validateColumnArgumentsV2(
	mapping schemav2.IdentifierMapping,
	arguments schemav2.ColumnArguments,
) error {
	reference := schemav2.FieldReference{
		FieldID: arguments.FieldID, Generation: arguments.Generation,
		Variant: arguments.Variant,
	}
	if err := requireFieldV2(mapping, reference); err != nil {
		return err
	}
	if arguments.Variant != schemav2.VariantValue &&
		arguments.StorageType != schemav2.StorageText {
		return fmt.Errorf("companion field storage must be text")
	}
	return validateColumnStorageV2(
		arguments.StorageType, arguments.DecimalPrecision, arguments.DecimalScale,
	)
}

func validateColumnStorageV2(
	storageType schemav2.StorageType,
	decimalPrecision *int,
	decimalScale *int,
) error {
	if !supportedStorageTypesV2[storageType] {
		return fmt.Errorf("storage type is not supported")
	}
	if storageType != schemav2.StorageDecimal {
		if decimalPrecision != nil || decimalScale != nil {
			return fmt.Errorf("decimal bounds are only valid for decimal storage")
		}
		return nil
	}
	if decimalPrecision == nil || decimalScale == nil ||
		*decimalPrecision < 1 || *decimalPrecision > 1000 ||
		*decimalScale < 0 || *decimalScale > *decimalPrecision {
		return fmt.Errorf("decimal precision or scale is invalid")
	}
	return nil
}
