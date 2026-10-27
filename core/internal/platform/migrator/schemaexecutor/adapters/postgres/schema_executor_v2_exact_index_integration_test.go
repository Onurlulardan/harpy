//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func TestSchemaExecutorV2RejectsEveryExtraPhysicalIndexOnPostgreSQL184(t *testing.T) {
	for _, variant := range []string{"fresh", "v1_adopted"} {
		t.Run(variant, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			environment := startSchemaExecutorIntegrationVariant(t, ctx, variant)
			environment.createTenant(t, ctx, v2ExactTenant, "tenant_v2_exact")
			plan := v2ExactPlan(t)
			target := plan.Steps[5].IdentifierMapping.RelatedForms[0].PhysicalName
			_, err := environment.pool.Exec(ctx, `CREATE TABLE tenant_v2_exact.`+
				pgIdentifier(target)+` (id uuid PRIMARY KEY)`)
			if err != nil {
				t.Fatal(err)
			}
			for ordinal := 0; ordinal < 5; ordinal++ {
				step := validatedV2Step(t, plan, ordinal)
				renderContext := ddl.RenderContext{
					TenantSchema: "tenant_v2_exact", TrigramSchema: "pg_catalog",
				}
				statement, renderErr := ddl.Render(renderContext, step)
				if renderErr != nil {
					t.Fatal(renderErr)
				}
				if _, err = environment.pool.Exec(ctx, statement.Text, statement.Arguments...); err != nil {
					t.Fatalf("apply V2 step %d: %v", ordinal, err)
				}
				acl, renderErr := ddl.RenderRecordACL(renderContext, step)
				if renderErr != nil {
					t.Fatal(renderErr)
				}
				for _, statement = range acl {
					if _, err = environment.pool.Exec(
						ctx, statement.Text, statement.Arguments...,
					); err != nil {
						t.Fatalf("apply V2 ACL step %d: %v", ordinal, err)
					}
				}
			}
			verify := validatedV2Step(t, plan, 5)
			if err = schemaExecutorVerifyStorageV2(ctx, environment.pool,
				"tenant_v2_exact", "pg_catalog", verify); err != nil {
				t.Fatalf("baseline V2 verification: %v", err)
			}
			form := pgIdentifier(plan.Steps[5].IdentifierMapping.Form.PhysicalName)
			field := pgIdentifier(plan.Steps[5].IdentifierMapping.Fields[0].PhysicalName)
			relation := pgIdentifier(plan.Steps[5].IdentifierMapping.Relations[0].PhysicalName)
			assertExtraIndexRejected(t, ctx, environment, verify,
				`CREATE INDEX extra_form_idx ON tenant_v2_exact.`+form+` (`+field+`)`,
				`DROP INDEX tenant_v2_exact.extra_form_idx`)
			assertExtraIndexRejected(t, ctx, environment, verify,
				`ALTER TABLE tenant_v2_exact.`+form+` ADD CONSTRAINT extra_unique UNIQUE (`+field+`)`,
				`ALTER TABLE tenant_v2_exact.`+form+` DROP CONSTRAINT extra_unique`)
			assertExtraIndexRejected(t, ctx, environment, verify,
				`CREATE INDEX extra_relation_idx ON tenant_v2_exact.`+relation+` (target_id)`,
				`DROP INDEX tenant_v2_exact.extra_relation_idx`)
		})
	}
}

func assertExtraIndexRejected(t *testing.T, ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	verify schemaexecutor.ValidatedStep, createSQL, dropSQL string) {
	t.Helper()
	if _, err := environment.pool.Exec(ctx, createSQL); err != nil {
		t.Fatal(err)
	}
	if err := schemaExecutorVerifyStorageV2(ctx, environment.pool,
		"tenant_v2_exact", "pg_catalog", verify); err == nil {
		t.Fatalf("extra physical index was accepted: %s", createSQL)
	}
	if _, err := environment.pool.Exec(ctx, dropSQL); err != nil {
		t.Fatal(err)
	}
}
