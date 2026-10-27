package postgres

import (
	"context"
	"slices"
)

func schemaExecutorVerifyFormConstraintsV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
	expectedConstraintCount int64,
) error {
	if expectedConstraintCount >= 0 {
		count, err := schemaExecutorTableCount(
			ctx, querier, schemaExecutorReadTableConstraintCountSQL, schemaName, tableName,
		)
		if err != nil {
			return err
		}
		if count != expectedConstraintCount {
			return schemaExecutorBlocked("schema_form_constraint_count_mismatch", nil)
		}
	}
	primary, found, err := schemaExecutorReadKey(
		ctx, querier, schemaName, tableName, tableName+"_pkey",
	)
	if err != nil {
		return err
	}
	if !found || primary.owner != "harpy_owner" || primary.kind != "p" ||
		!primary.validated || primary.deferrable || primary.deferred ||
		!slices.Equal(primary.sourceFields, []string{"id"}) {
		return schemaExecutorBlocked("schema_form_primary_key_mismatch", nil)
	}
	checks := map[string]string{
		"_row_version_check": "checkrow_version>0",
		"_id_v7_check":       "checknotuuid_extract_versionidisdistinctfrom7",
		"_created_by_check":  "checkcreated_by<>'00000000-0000-0000-0000-000000000000'::uuid",
		"_updated_by_check":  "checkupdated_by<>'00000000-0000-0000-0000-000000000000'::uuid",
		"_deleted_by_check":  "checkdeleted_byisnullordeleted_by<>'00000000-0000-0000-0000-000000000000'::uuid",
		"_timestamps_check": "checkisfinitecreated_atandisfiniteupdated_atandcreated_at<=updated_at" +
			"anddeleted_atisnullorisfinitedeleted_atanddeleted_at>=created_atanddeleted_at<=updated_at",
		"_soft_delete_check": "checkdeleted_atisnull=deleted_byisnull",
	}
	for suffix, definition := range checks {
		actual, exists, err := schemaExecutorReadConstraint(
			ctx, querier, schemaName, tableName, tableName+suffix,
		)
		if err != nil {
			return err
		}
		if !exists || actual.owner != "harpy_owner" || actual.kind != "c" ||
			!actual.validated || schemaExecutorNormalizeDefinition(actual.definition) != definition {
			return schemaExecutorBlocked("schema_form_hard_guard_mismatch", nil)
		}
	}
	return nil
}
