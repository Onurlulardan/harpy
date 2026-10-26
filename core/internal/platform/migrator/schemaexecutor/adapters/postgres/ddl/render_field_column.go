package ddl

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func renderFieldColumn(
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (Statement, error) {
	if step.ContractVersion() == schemav2.PlanContractVersion {
		return renderFieldColumnV2(schemaName, step)
	}
	mapping := step.Mapping()
	table := qualify(schemaName, mapping.Form.PhysicalName)
	if step.Operation() == schemav1.DropFieldColumn {
		arguments, ok := step.ArgumentValue().(*schemav1.FieldArguments)
		if !ok {
			return Statement{}, fmt.Errorf("field drop arguments are invalid")
		}
		column, err := fieldName(mapping, arguments.FieldID)
		if err != nil {
			return Statement{}, err
		}
		return Statement{Text: "ALTER TABLE " + table + " DROP COLUMN " + quote(column)}, nil
	}
	arguments, ok := step.ArgumentValue().(*schemav1.ColumnArguments)
	if !ok {
		return Statement{}, fmt.Errorf("field add arguments are invalid")
	}
	column, err := fieldName(mapping, arguments.FieldID)
	if err != nil {
		return Statement{}, err
	}
	columnType, err := ColumnType(*arguments)
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

func ColumnType(arguments schemav1.ColumnArguments) (string, error) {
	switch arguments.StorageType {
	case schemav1.StorageText, schemav1.StorageEnum:
		return "text", nil
	case schemav1.StorageInteger:
		return "bigint", nil
	case schemav1.StorageDecimal:
		return fmt.Sprintf("numeric(%d,%d)", *arguments.DecimalPrecision, *arguments.DecimalScale), nil
	case schemav1.StorageBoolean:
		return "boolean", nil
	case schemav1.StorageDate:
		return "date", nil
	case schemav1.StorageTimestamp:
		return "timestamp with time zone", nil
	case schemav1.StorageUUID, schemav1.StorageRelation:
		return "uuid", nil
	case schemav1.StorageTextList:
		return "text[]", nil
	case schemav1.StorageUUIDList:
		return "uuid[]", nil
	case schemav1.StorageJSON:
		return "jsonb", nil
	default:
		return "", fmt.Errorf("storage type is not renderable")
	}
}
