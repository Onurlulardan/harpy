package postgres

import (
	"context"
	"strings"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func schemaExecutorValidateBackfillColumns(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	step schemaexecutor.ValidatedStep,
) error {
	if step.ContractVersion() == schemav2.PlanContractVersion {
		return schemaExecutorValidateBackfillColumnsV2(
			ctx, querier, schemaName, step,
		)
	}
	arguments := step.ArgumentValue().(*schemav1.BackfillArguments)
	mapping := step.Mapping()
	sourceName, sourceMapped := schemaExecutorFieldGenerationName(
		mapping, arguments.SourceFieldID, arguments.SourceGeneration,
	)
	targetName, targetMapped := schemaExecutorFieldGenerationName(
		mapping, arguments.TargetFieldID, arguments.TargetGeneration,
	)
	if !sourceMapped || !targetMapped {
		return schemaExecutorBlocked("schema_backfill_mapping_missing", nil)
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
		!schemaExecutorConverterTypes(arguments.Converter, source.typeName, target.typeName) {
		return schemaExecutorBlocked("schema_backfill_type_mismatch", nil)
	}
	return nil
}

func schemaExecutorFieldGenerationName(
	mapping schemav1.IdentifierMapping,
	fieldID string,
	generation int64,
) (string, bool) {
	for _, field := range mapping.Fields {
		if field.LogicalID == fieldID && field.Generation == generation {
			return field.PhysicalName, true
		}
	}
	return "", false
}

func schemaExecutorConverterTypes(
	converter schemav1.Converter,
	source string,
	target string,
) bool {
	switch converter {
	case schemav1.ConvertIdentity:
		return source == target
	case schemav1.ConvertIntegerDecimal:
		return source == "bigint" && strings.HasPrefix(target, "numeric(")
	case schemav1.ConvertIntegerText:
		return source == "bigint" && target == "text"
	case schemav1.ConvertDecimalText:
		return strings.HasPrefix(source, "numeric(") && target == "text"
	case schemav1.ConvertDateTimestamp:
		return source == "date" && target == "timestamp with time zone"
	case schemav1.ConvertUUIDText:
		return source == "uuid" && target == "text"
	default:
		return false
	}
}
