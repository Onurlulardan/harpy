//go:build integration

package postgres

import (
	"encoding/json"
	"sort"
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func schemaExecutorArguments(t *testing.T, value any) schemav1.OperationArguments {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var result schemav1.OperationArguments
	if err := json.Unmarshal(document, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func schemaExecutorCreatePlan(
	t *testing.T,
	tenantID string,
	formID string,
	revision int64,
) schemav1.Plan {
	t.Helper()
	mapping := schemaExecutorMapping(t, formID, []schemav1.FieldIdentifier{},
		[]schemav1.Identifier{}, []schemav1.Identifier{})
	return schemaExecutorPlan(t, tenantID, formID, revision,
		[]schemav1.OperationID{schemav1.CreateFormTable}, []schemav1.Step{{
			Ordinal: 0, OperationID: schemav1.CreateFormTable,
			TransactionClass: schemav1.Transactional, ExecutorKind: schemav1.MigratorExecutor,
			IdentifierMapping: mapping, Arguments: schemav1.OperationArguments{},
		}})
}

func schemaExecutorCreateAndFieldPlan(
	t *testing.T,
	tenantID string,
	formID string,
	fieldID string,
) schemav1.Plan {
	t.Helper()
	field := schemav1.FieldIdentifier{
		LogicalID: fieldID, Generation: 1,
		PhysicalName: schemaExecutorFieldNameForTest(t, fieldID, 1),
	}
	mapping := schemaExecutorMapping(t, formID, []schemav1.FieldIdentifier{field},
		[]schemav1.Identifier{}, []schemav1.Identifier{})
	return schemaExecutorPlan(t, tenantID, formID, 1,
		[]schemav1.OperationID{schemav1.AddFieldColumn, schemav1.CreateFormTable},
		[]schemav1.Step{
			{Ordinal: 0, OperationID: schemav1.CreateFormTable,
				TransactionClass:  schemav1.Transactional,
				ExecutorKind:      schemav1.MigratorExecutor,
				IdentifierMapping: mapping, Arguments: schemav1.OperationArguments{}},
			{Ordinal: 1, OperationID: schemav1.AddFieldColumn,
				TransactionClass:  schemav1.Transactional,
				ExecutorKind:      schemav1.MigratorExecutor,
				IdentifierMapping: mapping,
				Arguments: schemaExecutorArguments(t, schemav1.ColumnArguments{
					FieldID: fieldID, StorageType: schemav1.StorageText, Nullable: true,
				})},
		})
}

func schemaExecutorBackfillPlan(
	t *testing.T,
	tenantID string,
	formID string,
	fieldID string,
) schemav1.Plan {
	t.Helper()
	fields := []schemav1.FieldIdentifier{
		{LogicalID: fieldID, Generation: 1,
			PhysicalName: schemaExecutorFieldNameForTest(t, fieldID, 1)},
		{LogicalID: fieldID, Generation: 2,
			PhysicalName: schemaExecutorFieldNameForTest(t, fieldID, 2)},
	}
	mapping := schemaExecutorMapping(t, formID, fields,
		[]schemav1.Identifier{}, []schemav1.Identifier{})
	arguments := schemav1.BackfillArguments{
		SourceFieldID: fieldID, SourceGeneration: 1,
		TargetFieldID: fieldID, TargetGeneration: 2,
		Converter: schemav1.ConvertIdentity,
	}
	return schemaExecutorPlan(t, tenantID, formID, 1,
		[]schemav1.OperationID{schemav1.BackfillFieldBatch}, []schemav1.Step{{
			Ordinal: 0, OperationID: schemav1.BackfillFieldBatch,
			TransactionClass: schemav1.Transactional, ExecutorKind: schemav1.MigratorExecutor,
			IdentifierMapping: mapping, Arguments: schemaExecutorArguments(t, arguments),
		}})
}

func schemaExecutorIndexPlan(
	t *testing.T,
	tenantID string,
	formID string,
	fieldID string,
	indexID string,
	unique bool,
) schemav1.Plan {
	t.Helper()
	field := schemav1.FieldIdentifier{
		LogicalID: fieldID, Generation: 1,
		PhysicalName: schemaExecutorFieldNameForTest(t, fieldID, 1),
	}
	index := schemav1.Identifier{
		LogicalID: indexID, PhysicalName: schemaExecutorObjectName(t, "idx_", indexID),
	}
	mapping := schemaExecutorMapping(t, formID, []schemav1.FieldIdentifier{field},
		[]schemav1.Identifier{index}, []schemav1.Identifier{})
	arguments := schemav1.IndexArguments{
		IndexID: indexID, FieldIDs: []string{fieldID}, Method: schemav1.IndexBTree,
		Unique: unique, ActiveRecordsOnly: true,
	}
	allowed := []schemav1.OperationID{schemav1.CreateIndexConcurrent}
	sort.Slice(allowed, func(i, j int) bool { return allowed[i] < allowed[j] })
	return schemaExecutorPlan(t, tenantID, formID, 1, allowed, []schemav1.Step{{
		Ordinal: 0, OperationID: schemav1.CreateIndexConcurrent,
		TransactionClass:  schemav1.ConcurrentIndex,
		ExecutorKind:      schemav1.MigratorExecutor,
		IdentifierMapping: mapping, Arguments: schemaExecutorArguments(t, arguments),
	}})
}

func schemaExecutorConstraintPlan(
	t *testing.T,
	tenantID string,
	formID string,
	fieldID string,
	constraintID string,
) schemav1.Plan {
	t.Helper()
	field := schemav1.FieldIdentifier{
		LogicalID: fieldID, Generation: 1,
		PhysicalName: schemaExecutorFieldNameForTest(t, fieldID, 1),
	}
	constraint := schemav1.Identifier{
		LogicalID:    constraintID,
		PhysicalName: schemaExecutorObjectName(t, "cst_", constraintID),
	}
	mapping := schemaExecutorMapping(t, formID,
		[]schemav1.FieldIdentifier{field}, []schemav1.Identifier{},
		[]schemav1.Identifier{constraint})
	arguments := schemav1.ConstraintArguments{
		ConstraintID: constraintID, Kind: schemav1.ConstraintNotNull,
		FieldIDs: []string{fieldID},
	}
	return schemaExecutorPlan(t, tenantID, formID, 1,
		[]schemav1.OperationID{schemav1.ValidateConstraint}, []schemav1.Step{{
			Ordinal: 0, OperationID: schemav1.ValidateConstraint,
			TransactionClass:  schemav1.Transactional,
			ExecutorKind:      schemav1.MigratorExecutor,
			IdentifierMapping: mapping,
			Arguments:         schemaExecutorArguments(t, arguments),
		}})
}
