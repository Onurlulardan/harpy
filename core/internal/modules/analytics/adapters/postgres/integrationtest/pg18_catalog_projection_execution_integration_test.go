//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"
)

const analyticsCatalogProjectionEpoch = "2026-07-30T20:00:00.000000Z"

type analyticsCatalogProjectionResult struct {
	receiptID, lifecycle string
	readyAt              sql.NullTime
}

func installFormsCatalogDependencyFixture(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
CREATE TABLE t_analytics_a.forms_catalog_dependency_fixture(
 slug text PRIMARY KEY,publication_id uuid NOT NULL,definition_id uuid NOT NULL,
 version_id uuid NOT NULL,owner_generation bigint NOT NULL,catalog_version bigint NOT NULL,
 query_contract_version bigint NOT NULL,policy_version bigint NOT NULL,
 data_revision_capability text NOT NULL,dependency_sha256 bytea NOT NULL
);
CREATE FUNCTION t_analytics_a.forms_catalog_analytics_dependency_v1(
 p_scope_schema name,p_slug text
) RETURNS TABLE(
 publication_id uuid,definition_id uuid,version_id uuid,owner_generation bigint,
 catalog_version bigint,query_contract_version bigint,policy_version bigint,
 data_revision_capability text,dependency_sha256 bytea
) LANGUAGE sql SECURITY DEFINER SET search_path TO pg_catalog AS $function$
 SELECT source.publication_id,source.definition_id,source.version_id,
 source.owner_generation,source.catalog_version,source.query_contract_version,
 source.policy_version,source.data_revision_capability,source.dependency_sha256
 FROM t_analytics_a.forms_catalog_dependency_fixture AS source
 WHERE p_scope_schema='t_analytics_a'::name AND source.slug=p_slug
$function$;
REVOKE ALL ON FUNCTION t_analytics_a.forms_catalog_analytics_dependency_v1(name,text)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION t_analytics_a.forms_catalog_analytics_dependency_v1(name,text)
TO harpy_runtime;`)
	if err != nil {
		t.Fatalf("install Forms dependency fixture: %v", err)
	}
}

func seedFormsCatalogDependency(
	t *testing.T, ctx context.Context, database *sql.DB, slug string,
) {
	t.Helper()
	var publicationID, definitionID, versionID, digestPair string
	switch slug {
	case "hr-employees":
		publicationID = "019f9000-0000-7000-8000-000000000031"
		definitionID = "019f9000-0000-7000-8000-000000000032"
		versionID = "019f9000-0000-7000-8000-000000000033"
		digestPair = "66"
	case "hr-departments":
		publicationID = "019f9000-0000-7000-8000-000000000034"
		definitionID = "019f9000-0000-7000-8000-000000000035"
		versionID = "019f9000-0000-7000-8000-000000000036"
		digestPair = "77"
	default:
		t.Fatalf("unknown Forms dependency fixture %q", slug)
	}
	_, err := database.ExecContext(ctx, `INSERT INTO
t_analytics_a.forms_catalog_dependency_fixture VALUES(
$1,$2,$3,$4,1,1,1,1,'forms.records.revision.v1',
decode(repeat($5,32),'hex'))`,
		slug, publicationID, definitionID, versionID, digestPair)
	if err != nil {
		t.Fatalf("seed Forms dependency fixture: %v", err)
	}
}

func projectAnalyticsCatalogFixture(
	ctx context.Context, connection *sql.Conn,
	fixture analyticsCatalogProjectionFixture,
	receiptID, releaseID, key string, occurredAt time.Time,
) (analyticsCatalogProjectionResult, error) {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return analyticsCatalogProjectionResult{}, err
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		return analyticsCatalogProjectionResult{}, err
	}
	requestSHA := sha256.Sum256([]byte(key))
	componentKind := "ENDPOINT"
	if fixture.kind == "dashboard" {
		componentKind = "DASHBOARD"
	}
	var result analyticsCatalogProjectionResult
	err = transaction.QueryRowContext(ctx, `SELECT receipt_id::text,lifecycle_state,ready_at
FROM analytics_catalog_project_v1(
't_analytics_a',$1,$2,'hr','1.0.0',$3,$4,1,
'analytics.catalog-definition.v1',$5,$6,'hr',$7,$8,1,$9,$10,$11,$12,$13,$14,$15,$16
)`, receiptID, analyticsCatalogPackageID, releaseID, fixture.componentID,
		fixture.contentSHA, componentKind, fixture.slug, fixture.semanticSHA[:],
		fixture.auxiliaryCount, fixture.document, fixture.runtimePlan, fixture.planSHA[:],
		analyticsCatalogActorID, key, requestSHA[:], occurredAt,
	).Scan(&result.receiptID, &result.lifecycle, &result.readyAt)
	if err != nil {
		return analyticsCatalogProjectionResult{}, err
	}
	return result, transaction.Commit()
}
