package ddl

import (
	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func renderFormTable(
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (Statement, error) {
	if step.ContractVersion() == schemav2.PlanContractVersion {
		return renderFormTableV2(schemaName, step)
	}
	table := qualify(schemaName, step.Mapping().Form.PhysicalName)
	if step.Operation() == schemav1.DropFormTable {
		return Statement{Text: "DROP TABLE " + table}, nil
	}
	return Statement{Text: "CREATE TABLE " + table + ` (
    id uuid PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    row_version bigint NOT NULL DEFAULT 1 CHECK (row_version > 0),
    deleted_at timestamptz
)`}, nil
}
