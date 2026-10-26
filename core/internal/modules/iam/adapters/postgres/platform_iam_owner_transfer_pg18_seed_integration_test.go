//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const ownerTransferAuthoritySeed = `
INSERT INTO platform.iam_platform_subjects (
    id,environment_id,primary_tenant_id,user_id,state,revision,created_at,updated_at,retired_at
) VALUES (
    '51000000-0000-4000-8000-000000000001','30000000-0000-4000-8000-000000000001',
    '10000000-0000-4000-8000-000000000001','41000000-0000-4000-8000-000000000001',
    'active',1,clock_timestamp(),clock_timestamp(),NULL
);
INSERT INTO platform.iam_platform_roles (
    id,environment_id,role_kind,role_key,state,revision,created_at,updated_at,retired_at
) VALUES (
    '52000000-0000-4000-8000-000000000001','30000000-0000-4000-8000-000000000001',
    'instance_owner','instance_owner','active',1,clock_timestamp(),clock_timestamp(),NULL
);
INSERT INTO platform.iam_platform_role_permissions (
    id,environment_id,role_id,permission_id,state,revision,created_at,updated_at,revoked_at
) SELECT
    '54000000-0000-4000-8000-000000000001','30000000-0000-4000-8000-000000000001',
    '52000000-0000-4000-8000-000000000001',permission.id,
    'active',1,clock_timestamp(),clock_timestamp(),NULL
FROM platform.iam_platform_permissions AS permission
WHERE permission.permission_key = 'environment_all';
INSERT INTO platform.iam_platform_subject_roles (
    id,environment_id,subject_id,role_id,state,revision,created_at,updated_at,revoked_at
) VALUES (
    '53000000-0000-4000-8000-000000000001','30000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000001','52000000-0000-4000-8000-000000000001',
    'active',1,clock_timestamp(),clock_timestamp(),NULL
);
INSERT INTO platform.iam_platform_owner_policies (
    environment_id,owner_binding_id,revision,created_at,updated_at
) VALUES (
    '30000000-0000-4000-8000-000000000001','53000000-0000-4000-8000-000000000001',
    1,clock_timestamp(),clock_timestamp()
);
INSERT INTO platform.iam_platform_authorization_clocks (
    environment_id,authorization_clock,revision,created_at,updated_at
) VALUES ('30000000-0000-4000-8000-000000000001',1,1,clock_timestamp(),clock_timestamp());
SELECT * FROM platform.iam_protected_authority_head_initialize_v1(
    '30000000-0000-4000-8000-000000000001',
    '55000000-0000-4000-8000-000000000001',1,
    decode(repeat('11',32),'hex'),decode(repeat('21',32),'hex')
);`

func seedOwnerTransferAuthority(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	if _, err := connection.ExecContext(ctx, ownerTransferAuthoritySeed); err != nil {
		t.Fatalf("seed owner transfer authority: %v", err)
	}
}
