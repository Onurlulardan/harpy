package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

var supportedStorageTypes = map[schemav1.StorageType]bool{
	schemav1.StorageText: true, schemav1.StorageInteger: true,
	schemav1.StorageDecimal: true, schemav1.StorageBoolean: true,
	schemav1.StorageDate: true, schemav1.StorageTimestamp: true,
	schemav1.StorageUUID: true, schemav1.StorageEnum: true,
	schemav1.StorageRelation: true, schemav1.StorageTextList: true,
	schemav1.StorageUUIDList: true, schemav1.StorageJSON: true,
}

func validateColumnArguments(
	mapping schemav1.IdentifierMapping,
	arguments schemav1.ColumnArguments,
) error {
	if err := requireField(mapping, arguments.FieldID); err != nil {
		return err
	}
	return validateColumnStorage(
		arguments.StorageType, arguments.DecimalPrecision, arguments.DecimalScale,
	)
}

func validateColumnStorage(
	storageType schemav1.StorageType,
	decimalPrecision *int,
	decimalScale *int,
) error {
	if !supportedStorageTypes[storageType] {
		return fmt.Errorf("storage type is not supported")
	}
	if storageType != schemav1.StorageDecimal {
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
