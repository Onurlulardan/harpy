package postgres

import (
	"context"
	"strings"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorValidateBackfillColumnsV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	step schemaexecutor.ValidatedStep,
) error {
	arguments, ok := step.ArgumentValue().(*schemav2.BackfillArguments)
	if !ok {
		return schemaExecutorBlocked("schema_backfill_contract_invalid", nil)
	}
	mapping, err := schemaExecutorMappingV2(step)
	if err != nil {
		return err
	}
	sourceName, err := schemaExecutorFieldNameV2(mapping, arguments.Source)
	if err != nil {
		return schemaExecutorBlocked("schema_backfill_mapping_missing", err)
	}
	targetName, err := schemaExecutorFieldNameV2(mapping, arguments.Target)
	if err != nil {
		return schemaExecutorBlocked("schema_backfill_mapping_missing", err)
	}
	source, sourceExists, err := schemaExecutorReadColumn(
		ctx, querier, schemaName, mapping.Form.PhysicalName, sourceName,
	)
	if err != nil {
		return err
	}
	target, targetExists, err := schemaExecutorReadColumn(
		ctx, querier, schemaName, mapping.Form.PhysicalName, targetName,
	)
	if err != nil {
		return err
	}
	if !sourceExists || !targetExists ||
		!schemaExecutorConverterTypesV2(arguments.Converter, source.typeName, target.typeName) {
		return schemaExecutorBlocked("schema_backfill_type_mismatch", nil)
	}
	if arguments.DefaultLiteral != nil && !schemaExecutorLiteralColumnV2(
		arguments.DefaultLiteral.StorageType, target.typeName,
	) {
		return schemaExecutorBlocked("schema_backfill_default_type_mismatch", nil)
	}
	return nil
}

func schemaExecutorLiteralColumnV2(storage schemav2.StorageType, column string) bool {
	switch storage {
	case schemav2.StorageText, schemav2.StorageEnum:
		return column == "text"
	case schemav2.StorageInteger:
		return column == "bigint"
	case schemav2.StorageDecimal:
		return strings.HasPrefix(column, "numeric(")
	case schemav2.StorageBoolean:
		return column == "boolean"
	case schemav2.StorageDate:
		return column == "date"
	case schemav2.StorageTimestamp:
		return column == "timestamp with time zone"
	case schemav2.StorageUUID:
		return column == "uuid"
	case schemav2.StorageTextList:
		return column == "text[]"
	case schemav2.StorageUUIDList:
		return column == "uuid[]"
	case schemav2.StorageJSON:
		return column == "jsonb"
	default:
		return false
	}
}

func schemaExecutorConverterTypesV2(
	converter schemav2.Converter,
	source string,
	target string,
) bool {
	switch converter {
	case schemav2.ConvertIdentity:
		return source == target
	case schemav2.ConvertIntegerDecimal:
		return source == "bigint" && strings.HasPrefix(target, "numeric(")
	case schemav2.ConvertDecimalWiden:
		return strings.HasPrefix(source, "numeric(") &&
			strings.HasPrefix(target, "numeric(") && source != target
	case schemav2.ConvertIntegerText:
		return source == "bigint" && target == "text"
	case schemav2.ConvertDecimalText:
		return strings.HasPrefix(source, "numeric(") && target == "text"
	case schemav2.ConvertDateTimestamp:
		return source == "date" && target == "timestamp with time zone"
	case schemav2.ConvertUUIDText:
		return source == "uuid" && target == "text"
	default:
		return false
	}
}
