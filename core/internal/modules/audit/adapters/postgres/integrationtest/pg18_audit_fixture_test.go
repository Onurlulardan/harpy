//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func createLegacyAuditFixture(t *testing.T, ctx context.Context, connection *sql.Conn) {
	t.Helper()
	_, err := connection.ExecContext(ctx, `CREATE TABLE change_logs (
        id uuid PRIMARY KEY,tenant_id uuid,correlation_id varchar(64),span_id varchar(64),
        entity_type varchar(50) NOT NULL,entity_id varchar(100) NOT NULL,form_slug varchar(100),
        action varchar(20) NOT NULL,old_data jsonb,new_data jsonb,changed_fields jsonb,
        method varchar(10),path varchar(500),ip_address varchar(45),user_agent text,status_code integer,
        user_id uuid,user_name varchar(255),org_id uuid,reason text,metadata jsonb,created_at timestamp);
        INSERT INTO change_logs(id,entity_type,entity_id,action,old_data,new_data,changed_fields,
            user_id,org_id,correlation_id,ip_address,user_agent,created_at) VALUES
        ('10000000-0000-4000-8000-000000000001','user','user-1','update',
         '{"password":"old"}','{"password":"new"}',
         '[{"field":"password","old_value":"old","new_value":"new"}]',
         '20000000-0000-4000-8000-000000000001','30000000-0000-4000-8000-000000000001',
         '40000000-0000-4000-8000-000000000001','127.0.0.1','secret-agent','2026-07-01 10:00:00'),
        ('10000000-0000-4000-8000-000000000002','form_record','record-1','create',NULL,
         '{"token":"secret"}',NULL,NULL,NULL,'not-a-uuid','127.0.0.2','secret-agent','2026-07-01 11:00:00'),
        ('10000000-0000-4000-8000-000000000003','document','document-1','delete',
         '{"content":"secret"}',NULL,NULL,NULL,NULL,NULL,'127.0.0.3','secret-agent','2026-07-01 12:00:00')`)
	if err != nil {
		t.Fatal(err)
	}
}

func containsAll(value string, expected ...string) bool {
	for _, item := range expected {
		if !strings.Contains(value, item) {
			return false
		}
	}
	return true
}
