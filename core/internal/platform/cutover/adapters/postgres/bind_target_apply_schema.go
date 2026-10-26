package postgres

import (
	"context"
	_ "embed"
	"fmt"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
)

//go:embed sql/bind_target_apply_schema.sql
var bindTargetApplySchemaSQL string

func bindTargetApplySchema(
	ctx context.Context,
	tx pgx.Tx,
	relation cutoverv1.RelationMapping,
	prepared preparedBatch,
) error {
	var schemaOID, relationOID, procedureOID int64
	err := tx.QueryRow(ctx, bindTargetApplySchemaSQL,
		int64(relation.Target.SchemaOID), relation.Target.SchemaName,
		int64(relation.Target.RelationOID), relation.Target.RelationName,
		prepared.procedureSignature, int64(prepared.procedureOID),
	).Scan(&schemaOID, &relationOID, &procedureOID)
	if err != nil || schemaOID != int64(relation.Target.SchemaOID) ||
		relationOID != int64(relation.Target.RelationOID) ||
		procedureOID != int64(prepared.procedureOID) {
		return fmt.Errorf("bind cutover target apply schema")
	}
	return nil
}
