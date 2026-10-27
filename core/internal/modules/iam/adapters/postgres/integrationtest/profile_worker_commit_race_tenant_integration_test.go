//go:build integration

package integrationtest

import (
	"context"
	"testing"
)

const (
	publicationRaceOtherTenantID = "10000000-0000-4000-8000-000000000002"
	publicationRaceOtherSchema   = "t_profile_publication_other"
)

func preparePublicationRaceTenant(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	statements := []struct {
		query string
		args  []any
	}{
		{`CREATE SCHEMA t_profile_publication_other
AUTHORIZATION harpy_owner`, nil},
		{`INSERT INTO platform.tenants
VALUES($1,'profile-publication-other',$2,'ACTIVE',NULL)`,
			[]any{publicationRaceOtherTenantID, publicationRaceOtherSchema}},
		{`INSERT INTO platform.domain_cutovers VALUES
('tenant',$1,'iam','v2','v2',1),
('tenant',$1,'system_resource_layout','v2','v2',1)`,
			[]any{publicationRaceOtherTenantID}},
		{`INSERT INTO platform.cutover_groups VALUES
('tenant',$1,'iam_system_resources','v2','v2',1)`,
			[]any{publicationRaceOtherTenantID}},
		{`INSERT INTO platform.cutover_group_members VALUES
('tenant',$1,'iam_system_resources',1,'iam'),
('tenant',$1,'iam_system_resources',2,'system_resource_layout')`,
			[]any{publicationRaceOtherTenantID}},
		{`CREATE TABLE t_profile_publication_other.tenancy_schema_identity(
singleton boolean PRIMARY KEY,tenant_id uuid NOT NULL UNIQUE,
schema_name text NOT NULL,origin text NOT NULL,created_at timestamptz NOT NULL)`, nil},
		{`INSERT INTO t_profile_publication_other.tenancy_schema_identity
VALUES(true,$1,$2,'new',pg_catalog.now())`,
			[]any{publicationRaceOtherTenantID, publicationRaceOtherSchema}},
		{`CREATE TABLE
t_profile_publication_other.iam_authorization_operation_receipts(
operation_id uuid PRIMARY KEY)`, nil},
		{`GRANT USAGE ON SCHEMA t_profile_publication_other
TO harpy_runtime`, nil},
		{`GRANT SELECT ON t_profile_publication_other.tenancy_schema_identity,
t_profile_publication_other.iam_authorization_operation_receipts
TO harpy_runtime`, nil},
	}
	for _, statement := range statements {
		if _, err := db.admin.Exec(ctx, statement.query, statement.args...); err != nil {
			t.Fatalf("prepare publication race tenant: %v", err)
		}
	}
}
