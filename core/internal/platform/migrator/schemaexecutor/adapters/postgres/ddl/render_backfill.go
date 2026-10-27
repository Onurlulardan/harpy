package ddl

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func renderBackfill(
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (Statement, error) {
	if step.ContractVersion() == schemav2.PlanContractVersion {
		return renderBackfillV2(schemaName, step)
	}
	arguments, ok := step.ArgumentValue().(*schemav1.BackfillArguments)
	if !ok {
		return Statement{}, fmt.Errorf("backfill arguments are invalid")
	}
	return renderBackfillStatement(schemaName, step.Mapping(), *arguments)
}

func renderBackfillStatement(
	schemaName string,
	mapping schemav1.IdentifierMapping,
	arguments schemav1.BackfillArguments,
) (Statement, error) {
	source, err := fieldGenerationName(
		mapping, arguments.SourceFieldID, arguments.SourceGeneration,
	)
	if err != nil {
		return Statement{}, err
	}
	target, err := fieldGenerationName(
		mapping, arguments.TargetFieldID, arguments.TargetGeneration,
	)
	if err != nil {
		return Statement{}, err
	}
	conversion, err := renderConversion(arguments.Converter, "selected.source_value")
	if err != nil {
		return Statement{}, err
	}
	table := qualify(schemaName, mapping.Form.PhysicalName)
	text := "WITH selected AS MATERIALIZED (SELECT id, row_version, " + quote(source) +
		" AS source_value FROM " + table +
		" WHERE ($1::uuid IS NULL OR id > $1::uuid) ORDER BY id LIMIT 1000 FOR UPDATE), " +
		"updated AS (UPDATE " + table + " AS target SET " + quote(target) + " = " +
		conversion + ", updated_at = pg_catalog.statement_timestamp(), " +
		"row_version = target.row_version + 1 FROM selected WHERE target.id = selected.id " +
		"AND target.row_version = selected.row_version RETURNING target.id) " +
		"SELECT (SELECT id::text FROM selected ORDER BY id DESC LIMIT 1), " +
		"(SELECT pg_catalog.count(*)::bigint FROM selected), " +
		"(SELECT id::text FROM updated ORDER BY id DESC LIMIT 1), " +
		"(SELECT pg_catalog.count(*)::bigint FROM updated)"
	return Statement{Text: text}, nil
}

func renderConversion(converter schemav1.Converter, value string) (string, error) {
	switch converter {
	case schemav1.ConvertIdentity:
		return value, nil
	case schemav1.ConvertIntegerDecimal:
		return value + "::numeric", nil
	case schemav1.ConvertIntegerText, schemav1.ConvertDecimalText,
		schemav1.ConvertUUIDText:
		return value + "::text", nil
	case schemav1.ConvertDateTimestamp:
		return value + "::timestamptz", nil
	default:
		return "", fmt.Errorf("converter is not renderable")
	}
}
