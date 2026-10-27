package ddl

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func renderFormTableV2(
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (Statement, error) {
	mapping, err := v2Mapping(step.V2Mapping)
	if err != nil {
		return Statement{}, err
	}
	table := qualify(schemaName, mapping.Form.PhysicalName)
	if step.Operation() == schemav1.DropFormTable {
		return Statement{Text: "DROP TABLE " + table}, nil
	}
	if err := validateEmptyArgumentsV2(step.ArgumentValue()); err != nil {
		return Statement{}, err
	}
	text := "CREATE TABLE " + table + ` (
    id uuid PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    created_by uuid NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    updated_by uuid NOT NULL,
    row_version bigint NOT NULL DEFAULT 1,
    deleted_at timestamptz,
    deleted_by uuid` + formTableChecksV2(mapping.Form.PhysicalName) + "\n)"
	return Statement{Text: text}, nil
}

func formTableChecksV2(tableName string) string {
	zeroUUID := "'00000000-0000-0000-0000-000000000000'::uuid"
	return `,
    CONSTRAINT ` + quote(tableName+"_row_version_check") + ` CHECK (row_version > 0),
    CONSTRAINT ` + quote(tableName+"_id_v7_check") +
		` CHECK (pg_catalog.uuid_extract_version(id) IS NOT DISTINCT FROM 7),
    CONSTRAINT ` + quote(tableName+"_created_by_check") +
		` CHECK (created_by <> ` + zeroUUID + `),
    CONSTRAINT ` + quote(tableName+"_updated_by_check") +
		` CHECK (updated_by <> ` + zeroUUID + `),
    CONSTRAINT ` + quote(tableName+"_deleted_by_check") +
		` CHECK (deleted_by IS NULL OR deleted_by <> ` + zeroUUID + `),
    CONSTRAINT ` + quote(tableName+"_timestamps_check") + ` CHECK (
        pg_catalog.isfinite(created_at) AND pg_catalog.isfinite(updated_at)
        AND created_at <= updated_at
        AND (deleted_at IS NULL OR (
			pg_catalog.isfinite(deleted_at) AND deleted_at >= created_at
			AND deleted_at <= updated_at
        ))
    ),
    CONSTRAINT ` + quote(tableName+"_soft_delete_check") +
		` CHECK ((deleted_at IS NULL) = (deleted_by IS NULL))`
}

func validateEmptyArgumentsV2(value any) error {
	if _, ok := value.(schemav2.EmptyArguments); !ok {
		return fmt.Errorf("V2 form table arguments are invalid")
	}
	return nil
}
