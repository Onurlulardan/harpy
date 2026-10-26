package ddl

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func renderFieldColumnV2(
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (Statement, error) {
	mapping, err := v2Mapping(step.V2Mapping)
	if err != nil {
		return Statement{}, err
	}
	table := qualify(schemaName, mapping.Form.PhysicalName)
	if step.Operation() == schemav1.OperationID(schemav2.SetFieldNotNull) {
		arguments, ok := step.ArgumentValue().(*schemav2.FieldArguments)
		if !ok {
			return Statement{}, fmt.Errorf("V2 not-null arguments are invalid")
		}
		column, err := fieldNameV2(mapping, schemav2.FieldReference{
			FieldID: arguments.FieldID, Generation: arguments.Generation,
			Variant: arguments.Variant,
		})
		if err != nil {
			return Statement{}, err
		}
		return Statement{Text: "ALTER TABLE " + table + " ALTER COLUMN " +
			quote(column) + " SET NOT NULL"}, nil
	}
	if step.Operation() == schemav1.DropFieldColumn {
		arguments, ok := step.ArgumentValue().(*schemav2.FieldArguments)
		if !ok {
			return Statement{}, fmt.Errorf("V2 field drop arguments are invalid")
		}
		column, err := fieldNameV2(mapping, schemav2.FieldReference{
			FieldID: arguments.FieldID, Generation: arguments.Generation,
			Variant: arguments.Variant,
		})
		if err != nil {
			return Statement{}, err
		}
		return Statement{Text: "ALTER TABLE " + table + " DROP COLUMN " + quote(column)}, nil
	}
	arguments, ok := step.ArgumentValue().(*schemav2.ColumnArguments)
	if !ok {
		return Statement{}, fmt.Errorf("V2 field add arguments are invalid")
	}
	column, err := fieldNameV2(mapping, schemav2.FieldReference{
		FieldID: arguments.FieldID, Generation: arguments.Generation,
		Variant: arguments.Variant,
	})
	if err != nil {
		return Statement{}, err
	}
	columnType, err := ColumnTypeV2(*arguments)
	if err != nil {
		return Statement{}, err
	}
	nullability := ""
	if !arguments.Nullable {
		nullability = " NOT NULL"
	}
	return Statement{Text: "ALTER TABLE " + table + " ADD COLUMN " +
		quote(column) + " " + columnType + nullability}, nil
}

func ColumnTypeV2(arguments schemav2.ColumnArguments) (string, error) {
	switch arguments.StorageType {
	case schemav2.StorageText, schemav2.StorageEnum:
		return "text", nil
	case schemav2.StorageInteger:
		return "bigint", nil
	case schemav2.StorageDecimal:
		return fmt.Sprintf("numeric(%d,%d)", *arguments.DecimalPrecision, *arguments.DecimalScale), nil
	case schemav2.StorageBoolean:
		return "boolean", nil
	case schemav2.StorageDate:
		return "date", nil
	case schemav2.StorageTimestamp:
		return "timestamp with time zone", nil
	case schemav2.StorageUUID:
		return "uuid", nil
	case schemav2.StorageTextList:
		return "text[]", nil
	case schemav2.StorageUUIDList:
		return "uuid[]", nil
	case schemav2.StorageJSON:
		return "jsonb", nil
	default:
		return "", fmt.Errorf("V2 storage type is not renderable")
	}
}
