//go:build integration

package integrationtest

import (
	"context"
	"testing"
)

func seedPublicationPublicDeleteAuthority(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	const sql = `
INSERT INTO iam_permissions(
 id,match_kind,resource_id,action_id,is_protected,state,revision,created_at,updated_at)
VALUES
($1,'exact','91000000-0000-4000-8000-000000000001',
 (SELECT id FROM iam_actions WHERE action_key='read'),false,'active',1,
 pg_catalog.now(),pg_catalog.now()),
($2,'exact','91000000-0000-4000-8000-000000000001',
 (SELECT id FROM iam_actions WHERE action_key='delete'),false,'active',1,
 pg_catalog.now(),pg_catalog.now());
INSERT INTO iam_user_permission_bindings(
 id,user_id,permission_id,permission_match_kind,state,revision,
 created_at,updated_at,activated_at)
VALUES
('a0040000-0000-4000-8000-000000000002',$3,$1,'exact','active',1,
 pg_catalog.now(),pg_catalog.now(),pg_catalog.now()),
('a0040000-0000-4000-8000-000000000003',$3,$2,'exact','active',1,
 pg_catalog.now(),pg_catalog.now(),pg_catalog.now());`
	execPublicationTenant(t, ctx, db, sql,
		publicDeleteReadPermID, publicDeleteDeletePermID, publicationActorID)
}

func seedPublicationPublicDeleteDirectory(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	const sql = `
INSERT INTO platform.iam_login_directory(
 tenant_id,identifier_kind,lookup_key_id,identifier_hmac,user_id,
 source_identifier_id,source_revision,state,source_sha256,projected_at)
VALUES($1,'email_v1','delete-public-key',
 pg_catalog.sha256(pg_catalog.convert_to('delete-public-hmac','UTF8')),$2,$3,1,
 'active',pg_catalog.sha256(pg_catalog.convert_to('delete-public-source','UTF8')),
 pg_catalog.now());`
	execPublicationTenant(t, ctx, db, sql,
		db.tenant.id, publicDeleteUserID, publicDeleteIdentifierID)
}
