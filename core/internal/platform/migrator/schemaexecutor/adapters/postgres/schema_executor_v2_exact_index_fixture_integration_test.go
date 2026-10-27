//go:build integration

package postgres

import (
	"encoding/json"
	"sort"
	"testing"
	"time"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

const (
	v2ExactTenant = "018f47f4-7b1b-7cc0-8000-000000000811"
	v2ExactForm   = "018f47f4-7b1b-7cc0-8000-000000000812"
	v2ExactTarget = "018f47f4-7b1b-7cc0-8000-000000000813"
	v2ExactField  = "018f47f4-7b1b-7cc0-8000-000000000814"
	v2ExactRel    = "018f47f4-7b1b-7cc0-8000-000000000815"
	v2ExactSrcFK  = "018f47f4-7b1b-7cc0-8000-000000000816"
	v2ExactDstFK  = "018f47f4-7b1b-7cc0-8000-000000000817"
	v2ExactIndex  = "018f47f4-7b1b-7cc0-8000-000000000818"
	v2ExactUnique = "018f47f4-7b1b-7cc0-8000-000000000819"
)

func v2ExactPlan(t *testing.T) schemav2.Plan {
	t.Helper()
	form := v2Identifier(t, schemav2.IdentifierForm, v2ExactForm)
	target := v2Identifier(t, schemav2.IdentifierForm, v2ExactTarget)
	field := schemav2.FieldIdentifier{LogicalID: v2ExactField, Generation: 1,
		Variant: schemav2.VariantValue, PhysicalName: v2FieldName(t, v2ExactField)}
	relation := v2Identifier(t, schemav2.IdentifierRelation, v2ExactRel)
	index := v2IndexIdentifier(t, v2ExactIndex, 1)
	sourceFK := v2Identifier(t, schemav2.IdentifierConstraint, v2ExactSrcFK)
	targetFK := v2Identifier(t, schemav2.IdentifierConstraint, v2ExactDstFK)
	unique := v2ConstraintIdentifier(t, v2ExactUnique, 1)
	base := v2Mapping(form, nil, nil, nil, nil, nil)
	fieldMap := v2Mapping(form, nil, []schemav2.FieldIdentifier{field}, nil, nil, nil)
	relationMap := v2Mapping(form, []schemav2.Identifier{target}, nil,
		[]schemav2.Identifier{relation}, nil, []schemav2.Identifier{sourceFK, targetFK})
	constraintMap := v2Mapping(form, nil, []schemav2.FieldIdentifier{field}, nil,
		nil, []schemav2.Identifier{unique})
	indexMap := v2Mapping(form, nil, []schemav2.FieldIdentifier{field}, nil,
		[]schemav2.Identifier{index}, nil)
	unionConstraints := []schemav2.Identifier{sourceFK, targetFK, unique}
	sort.Slice(unionConstraints, func(i, j int) bool {
		return unionConstraints[i].LogicalID < unionConstraints[j].LogicalID
	})
	union := v2Mapping(form, []schemav2.Identifier{target},
		[]schemav2.FieldIdentifier{field}, []schemav2.Identifier{relation},
		[]schemav2.Identifier{index}, unionConstraints)
	relationArgs := schemav2.RelationArguments{RelationID: v2ExactRel,
		TargetFormID: v2ExactTarget, SourceConstraintID: v2ExactSrcFK,
		TargetConstraintID: v2ExactDstFK, Cardinality: schemav2.CardinalityMany,
		DeletePolicy: schemav2.DeleteRestrict}
	indexArgs := schemav2.IndexArguments{IndexID: v2ExactIndex, Generation: 1,
		Fields: []schemav2.IndexField{{FieldID: v2ExactField, Generation: 1,
			Variant: schemav2.VariantValue, StorageType: schemav2.StorageText}},
		Method: schemav2.IndexBTree, OperatorClass: schemav2.OperatorClassDefault,
		ActiveRecordsOnly: true}
	constraintArgs := schemav2.ConstraintArguments{ConstraintID: v2ExactUnique,
		Generation: 1, Kind: schemav2.ConstraintUnique, Fields: []schemav2.FieldReference{{
			FieldID: v2ExactField, Generation: 1, Variant: schemav2.VariantValue}}}
	verify := schemav2.VerifyArguments{Checks: []schemav2.VerifyCheck{
		schemav2.VerifyConstraints, schemav2.VerifyFields, schemav2.VerifyFormTable,
		schemav2.VerifyIndexes, schemav2.VerifyRelations, schemav2.VerifySystemColumns},
		Fields: []schemav2.VerifyFieldContract{{FieldID: v2ExactField, Generation: 1,
			Variant: schemav2.VariantValue, StorageType: schemav2.StorageText,
			Nullable: true}}, Relations: []schemav2.RelationArguments{relationArgs},
		Indexes: []schemav2.VerifyIndexContract{{IndexID: v2ExactIndex, Generation: 1,
			Fields: indexArgs.Fields, Method: indexArgs.Method,
			OperatorClass: indexArgs.OperatorClass, ActiveRecordsOnly: true}},
		Evolution: []schemav2.EvolutionDataContract{},
		Constraints: []schemav2.VerifyConstraintContract{{ConstraintID: v2ExactUnique,
			Generation: 1, Kind: schemav2.ConstraintUnique, Fields: constraintArgs.Fields}}}
	steps := []schemav2.Step{
		v2Step(0, schemav2.CreateFormTable, base, schemav2.EmptyArguments{}),
		v2Step(1, schemav2.AddFieldColumn, fieldMap, schemav2.ColumnArguments{
			FieldID: v2ExactField, Generation: 1, Variant: schemav2.VariantValue,
			StorageType: schemav2.StorageText, Nullable: true}),
		v2Step(2, schemav2.AddRelationTable, relationMap, relationArgs),
		v2Step(3, schemav2.AddConstraint, constraintMap, constraintArgs),
		v2Step(4, schemav2.CreateIndex, indexMap, indexArgs),
		v2Step(5, schemav2.VerifyStorage, union, verify),
	}
	return schemav2.Plan{ContractVersion: schemav2.PlanContractVersion,
		TenantID: v2ExactTenant, Owner: "forms",
		Resource: schemav2.Resource{Kind: schemav2.ResourceFormStorage,
			ID: v2ExactForm, Revision: 1},
		Compiler: schemav2.Compiler{Version: schemav2.CompilerVersion,
			PostgreSQLMajor: schemav2.RequiredPostgreSQLMajor},
		AllowedOperations: []schemav2.OperationID{schemav2.AddConstraint,
			schemav2.AddFieldColumn, schemav2.AddRelationTable,
			schemav2.CreateFormTable, schemav2.CreateIndex, schemav2.VerifyStorage},
		Steps: steps}
}

func v2Step(ordinal int, operation schemav2.OperationID,
	mapping schemav2.IdentifierMapping, value any) schemav2.Step {
	document, _ := json.Marshal(value)
	var arguments schemav2.OperationArguments
	_ = json.Unmarshal(document, &arguments)
	return schemav2.Step{Ordinal: ordinal, OperationID: operation,
		TransactionClass: schemav2.Transactional, ExecutorKind: schemav2.MigratorExecutor,
		IdentifierMapping: mapping, Arguments: arguments}
}

func v2Mapping(form schemav2.Identifier, related []schemav2.Identifier,
	fields []schemav2.FieldIdentifier, relations, indexes,
	constraints []schemav2.Identifier) schemav2.IdentifierMapping {
	return schemav2.IdentifierMapping{Form: form, RelatedForms: nonNilV2(related),
		Fields: nonNilV2(fields), Relations: nonNilV2(relations),
		Indexes: nonNilV2(indexes), Constraints: nonNilV2(constraints)}
}

func nonNilV2[T any](values []T) []T {
	if values == nil {
		return []T{}
	}
	return values
}

func v2Identifier(t *testing.T, kind schemav2.IdentifierKind, id string) schemav2.Identifier {
	t.Helper()
	name, err := schemav2.DeriveIdentifier(kind, id)
	if err != nil {
		t.Fatal(err)
	}
	return schemav2.Identifier{LogicalID: id, PhysicalName: name}
}

func v2IndexIdentifier(t *testing.T, id string, generation int64) schemav2.Identifier {
	t.Helper()
	name, err := schemav2.DeriveIndexIdentifier(id, generation)
	if err != nil {
		t.Fatal(err)
	}
	return schemav2.Identifier{LogicalID: id, Generation: generation, PhysicalName: name}
}

func v2ConstraintIdentifier(t *testing.T, id string, generation int64) schemav2.Identifier {
	t.Helper()
	name, err := schemav2.DeriveConstraintIdentifier(id, generation)
	if err != nil {
		t.Fatal(err)
	}
	return schemav2.Identifier{LogicalID: id, Generation: generation, PhysicalName: name}
}

func v2FieldName(t *testing.T, id string) string {
	t.Helper()
	name, err := schemav2.DeriveFieldIdentifier(id, 1, schemav2.VariantValue)
	if err != nil {
		t.Fatal(err)
	}
	return name
}

func validatedV2Step(t *testing.T, plan schemav2.Plan, ordinal int) schemaexecutor.ValidatedStep {
	t.Helper()
	prepared, err := schemaexecutor.BuildPlanV2(plan)
	if err != nil {
		t.Fatal(err)
	}
	rows := make([]schemaexecutor.StepRecord, len(plan.Steps))
	for index, step := range plan.Steps {
		stored := prepared.Steps()[index]
		rows[index] = schemaexecutor.StepRecord{Ordinal: index, Generation: 1, Attempt: 1,
			OperationID: string(step.OperationID), TransactionClass: string(step.TransactionClass),
			ExecutorKind: string(step.ExecutorKind), IdentifierMapping: stored.IdentifierMappingJSON(),
			Arguments: stored.ArgumentsJSON(), StepSHA256: stored.SHA256()}
	}
	now := time.Now().UTC()
	jobID := schemaExecutorIntegrationUUID(881)
	record := schemaexecutor.ClaimRecord{Job: schemaexecutor.JobRecord{JobID: jobID,
		TenantID: plan.TenantID, Owner: plan.Owner, ResourceKind: string(plan.Resource.Kind),
		ResourceID: plan.Resource.ID, ResourceRevision: plan.Resource.Revision,
		CompilerVersion: plan.Compiler.Version, PlanContractVersion: plan.ContractVersion,
		CanonicalPlan: prepared.CanonicalJSON(), PlanSHA256: prepared.SHA256(),
		StepCount: len(rows), NextStepOrdinal: ordinal, Attempt: 1, Generation: 1}, Steps: rows}
	record.Lease = schemaexecutor.LeaseRecord{JobID: jobID, TenantID: plan.TenantID,
		LeaseID: schemaExecutorIntegrationUUID(882), ExecutorID: schemaExecutorIntegrationUUID(883),
		JobAttempt: 1, StepAttempt: 1, StepOrdinal: ordinal,
		PlanSHA256: prepared.SHA256(), StepSHA256: rows[ordinal].StepSHA256,
		StepStartedAt: now, LeaseExpiresAt: now.Add(time.Minute), JobGeneration: 1,
		StepGeneration: 1}
	_, validated, err := schemaexecutor.AcceptIntegrationClaim(record)
	if err != nil {
		t.Fatal(err)
	}
	step, ok := validated.StepAt(ordinal)
	if !ok {
		t.Fatal("V2 step missing")
	}
	return step
}
