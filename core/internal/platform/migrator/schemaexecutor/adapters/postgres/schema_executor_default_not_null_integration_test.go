//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
	"github.com/jackc/pgx/v5"
)

func TestSchemaExecutorV2DefaultBackfillSetsNotNullOnPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	environment := startSchemaExecutorIntegration(t, ctx)
	environment.createTenant(t, ctx, v2ExactTenant, "tenant_default")
	formName, err := schemav2.DeriveIdentifier(schemav2.IdentifierForm, v2ExactForm)
	if err != nil {
		t.Fatal(err)
	}
	sourceName, err := schemav2.DeriveFieldIdentifier(
		v2ExactField, 1, schemav2.VariantValue,
	)
	if err != nil {
		t.Fatal(err)
	}
	targetName, err := schemav2.DeriveFieldIdentifier(
		v2ExactField, 2, schemav2.VariantValue,
	)
	if err != nil {
		t.Fatal(err)
	}
	table := pgx.Identifier{"tenant_default", formName}.Sanitize()
	quotedSource := pgx.Identifier{sourceName}.Sanitize()
	quotedTarget := pgx.Identifier{targetName}.Sanitize()
	_, err = environment.pool.Exec(ctx, "CREATE TABLE "+table+
		" (id uuid PRIMARY KEY,row_version bigint NOT NULL,updated_at timestamptz NOT NULL,"+
		quotedSource+" text,"+quotedTarget+" text)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = environment.pool.Exec(ctx, "INSERT INTO "+table+
		" VALUES($1,1,pg_catalog.statement_timestamp(),NULL,NULL),"+
		"($2,1,pg_catalog.statement_timestamp(),$3,NULL)",
		schemaExecutorIntegrationUUID(901), schemaExecutorIntegrationUUID(902), "kept")
	if err != nil {
		t.Fatal(err)
	}
	form := schemav2.Identifier{LogicalID: v2ExactForm, PhysicalName: formName}
	fields := []schemav2.FieldIdentifier{
		{LogicalID: v2ExactField, Generation: 1,
			Variant: schemav2.VariantValue, PhysicalName: sourceName},
		{LogicalID: v2ExactField, Generation: 2,
			Variant: schemav2.VariantValue, PhysicalName: targetName},
	}
	mapping := v2Mapping(form, nil, fields, nil, nil, nil)
	literal := &schemav2.TypedLiteral{
		StorageType: schemav2.StorageText, Value: []byte(`"created"`),
	}
	plan := schemav2.Plan{ContractVersion: schemav2.PlanContractVersion,
		TenantID: v2ExactTenant, Owner: "forms",
		Resource: schemav2.Resource{Kind: schemav2.ResourceFormStorage,
			ID: v2ExactForm, Revision: 1},
		Compiler: schemav2.Compiler{Version: schemav2.CompilerVersion,
			PostgreSQLMajor: schemav2.RequiredPostgreSQLMajor},
		AllowedOperations: []schemav2.OperationID{
			schemav2.BackfillFieldBatch, schemav2.SetFieldNotNull},
		Steps: []schemav2.Step{
			v2Step(0, schemav2.BackfillFieldBatch, mapping,
				schemav2.BackfillArguments{Source: schemav2.FieldReference{
					FieldID: v2ExactField, Generation: 1, Variant: schemav2.VariantValue},
					Target: schemav2.FieldReference{FieldID: v2ExactField,
						Generation: 2, Variant: schemav2.VariantValue},
					Converter: schemav2.ConvertIdentity, DefaultLiteral: literal}),
			v2Step(1, schemav2.SetFieldNotNull, mapping,
				schemav2.FieldArguments{FieldID: v2ExactField,
					Generation: 2, Variant: schemav2.VariantValue}),
		}}
	backfill, err := ddl.Render(ddl.RenderContext{TenantSchema: "tenant_default"},
		validatedV2Step(t, plan, 0))
	if err != nil {
		t.Fatal(err)
	}
	var cursor, updatedCursor string
	var selected, updated, selectedBytes, updatedBytes, nextRowBytes int64
	var selectedSHA, updatedSHA string
	err = environment.pool.QueryRow(ctx, backfill.Text, nil).Scan(
		&cursor, &selected, &selectedBytes, &selectedSHA,
		&updatedCursor, &updated, &updatedBytes, &updatedSHA, &nextRowBytes,
	)
	if err != nil || selected != 2 || updated != 2 || selectedBytes < 1 ||
		selectedBytes != updatedBytes || selectedSHA != updatedSHA || nextRowBytes != 0 {
		t.Fatalf("backfill = %v/%d/%d bytes=%d/%d next=%d",
			err, selected, updated, selectedBytes, updatedBytes, nextRowBytes)
	}
	notNull, err := ddl.Render(ddl.RenderContext{TenantSchema: "tenant_default"},
		validatedV2Step(t, plan, 1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = environment.pool.Exec(ctx, notNull.Text); err != nil {
		t.Fatal(err)
	}
	var defaults, kept int
	err = environment.pool.QueryRow(ctx, "SELECT count(*) FILTER (WHERE "+quotedTarget+
		"='created'),count(*) FILTER (WHERE "+quotedTarget+"='kept') FROM "+table).Scan(
		&defaults, &kept)
	if err != nil || defaults != 1 || kept != 1 {
		t.Fatalf("backfilled values = %d/%d: %v", defaults, kept, err)
	}
}
