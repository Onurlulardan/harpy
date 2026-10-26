package ddl

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

const recordReaderRole = "harpy_forms_record_reader"
const recordWriterRole = "harpy_forms_record_writer"

func RenderRecordACL(
	renderContext RenderContext,
	step schemaexecutor.ValidatedStep,
) ([]Statement, error) {
	mapping, ok := step.V2Mapping()
	if !ok {
		return nil, nil
	}
	table, applies, err := recordACLTableV2(mapping, step)
	if err != nil || !applies {
		return nil, err
	}
	return renderRecordACLStatementsV2(renderContext.TenantSchema, table), nil
}

func recordACLTableV2(
	mapping schemav2.IdentifierMapping,
	step schemaexecutor.ValidatedStep,
) (string, bool, error) {
	switch step.Operation() {
	case schemav1.CreateFormTable:
		return mapping.Form.PhysicalName, true, nil
	case schemav1.AddRelationTable:
		arguments, ok := step.ArgumentValue().(*schemav2.RelationArguments)
		if !ok {
			return "", false, fmt.Errorf("V2 relation ACL arguments are invalid")
		}
		table, err := identifierNameV2(mapping.Relations, arguments.RelationID)
		return table, true, err
	default:
		return "", false, nil
	}
}

func renderRecordACLStatementsV2(schemaName, tableName string) []Statement {
	schema := quote(schemaName)
	table := qualify(schemaName, tableName)
	return []Statement{
		{Text: "REVOKE CREATE ON SCHEMA " + schema +
			" FROM PUBLIC, harpy_runtime, harpy_v2_runtime, " +
			recordReaderRole + ", " + recordWriterRole},
		{Text: "GRANT USAGE ON SCHEMA " + schema +
			" TO " + recordReaderRole + ", " + recordWriterRole},
		{Text: "REVOKE ALL PRIVILEGES ON TABLE " + table +
			" FROM PUBLIC, harpy_runtime, harpy_v2_runtime, " +
			recordReaderRole + ", " + recordWriterRole},
		{Text: "GRANT SELECT ON TABLE " + table + " TO " + recordReaderRole},
		{Text: "GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE " + table +
			" TO " + recordWriterRole},
	}
}
