//go:build integration

package postgres

import (
	"context"
	"testing"

	pagev1 "github.com/Onurlulardan/harpy/contracts/formsrecordpage/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

const (
	v2BackfillTenant = "018f47f4-7b1b-7cc0-8000-000000000921"
	v2BackfillForm   = "018f47f4-7b1b-7cc0-8000-000000000922"
	v2BackfillField  = "018f47f4-7b1b-7cc0-8000-000000000923"
)

func v2BackfillPlan(t *testing.T) schemav2.Plan {
	t.Helper()
	form := v2Identifier(t, schemav2.IdentifierForm, v2BackfillForm)
	sourceName, err := schemav2.DeriveFieldIdentifier(
		v2BackfillField, 1, schemav2.VariantValue,
	)
	if err != nil {
		t.Fatal(err)
	}
	targetName, err := schemav2.DeriveFieldIdentifier(
		v2BackfillField, 2, schemav2.VariantValue,
	)
	if err != nil {
		t.Fatal(err)
	}
	fields := []schemav2.FieldIdentifier{
		{LogicalID: v2BackfillField, Generation: 1,
			Variant: schemav2.VariantValue, PhysicalName: sourceName},
		{LogicalID: v2BackfillField, Generation: 2,
			Variant: schemav2.VariantValue, PhysicalName: targetName},
	}
	mapping := v2Mapping(form, nil, fields, nil, nil, nil)
	step := v2Step(0, schemav2.BackfillFieldBatch, mapping,
		schemav2.BackfillArguments{
			Source: schemav2.FieldReference{FieldID: v2BackfillField,
				Generation: 1, Variant: schemav2.VariantValue},
			Target: schemav2.FieldReference{FieldID: v2BackfillField,
				Generation: 2, Variant: schemav2.VariantValue},
			Converter: schemav2.ConvertIdentity,
		})
	return schemav2.Plan{ContractVersion: schemav2.PlanContractVersion,
		TenantID: v2BackfillTenant, Owner: "forms",
		Resource: schemav2.Resource{Kind: schemav2.ResourceFormStorage,
			ID: v2BackfillForm, Revision: 1},
		Compiler: schemav2.Compiler{Version: schemav2.CompilerVersion,
			PostgreSQLMajor: schemav2.RequiredPostgreSQLMajor},
		AllowedOperations: []schemav2.OperationID{schemav2.BackfillFieldBatch},
		Steps:             []schemav2.Step{step}}
}

func seedV2BackfillRows(
	t *testing.T, ctx context.Context, environment *schemaExecutorIntegrationEnvironment,
	plan schemav2.Plan,
) {
	t.Helper()
	mapping := plan.Steps[0].IdentifierMapping
	table := pgIdentifier("tenant_backfill_v2") + "." +
		pgIdentifier(mapping.Form.PhysicalName)
	source := pgIdentifier(mapping.Fields[0].PhysicalName)
	target := pgIdentifier(mapping.Fields[1].PhysicalName)
	_, err := environment.pool.Exec(ctx, "CREATE TABLE "+table+
		" (id uuid PRIMARY KEY,row_version bigint NOT NULL,"+
		"updated_at timestamptz NOT NULL,"+source+" text,"+target+" text)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = environment.pool.Exec(ctx, "INSERT INTO "+table+
		" (id,row_version,updated_at,"+source+") SELECT "+
		"('60000000-0000-4000-8000-'||pg_catalog.lpad("+
		"pg_catalog.to_hex(value),12,'0'))::uuid,1,"+
		"pg_catalog.statement_timestamp(),CASE WHEN value<=9 THEN "+
		"pg_catalog.repeat('x',1048000) ELSE 'small-'||value::text END "+
		"FROM pg_catalog.generate_series(1,1010) AS value")
	if err != nil {
		t.Fatal(err)
	}
}

func seedV2OversizedBackfillRow(
	t *testing.T, ctx context.Context, environment *schemaExecutorIntegrationEnvironment,
	plan schemav2.Plan,
) {
	t.Helper()
	mapping := plan.Steps[0].IdentifierMapping
	table := pgIdentifier("tenant_backfill_v2") + "." +
		pgIdentifier(mapping.Form.PhysicalName)
	source := pgIdentifier(mapping.Fields[0].PhysicalName)
	target := pgIdentifier(mapping.Fields[1].PhysicalName)
	_, err := environment.pool.Exec(ctx, "CREATE TABLE "+table+
		" (id uuid PRIMARY KEY,row_version bigint NOT NULL,"+
		"updated_at timestamptz NOT NULL,"+source+" text,"+target+" text)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = environment.pool.Exec(ctx, "INSERT INTO "+table+
		" (id,row_version,updated_at,"+source+") VALUES "+
		"('60000000-0000-4000-8000-000000000001',1,"+
		"pg_catalog.statement_timestamp(),'small-before-oversized'),"+
		"('60000000-0000-4000-8000-000000000002',1,"+
		"pg_catalog.statement_timestamp(),pg_catalog.repeat('x',$1))",
		pagev1.MaximumRowBytes+1)
	if err != nil {
		t.Fatal(err)
	}
}
