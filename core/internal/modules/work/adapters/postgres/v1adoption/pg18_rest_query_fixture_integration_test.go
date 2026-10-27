//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"testing"
)

const (
	workQueryPrincipal = "USER:31000000-0000-4000-8000-000000000001"
	workOtherPrincipal = "USER:31000000-0000-4000-8000-000000000002"
)

func seedWorkQueryFixture(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) {
	t.Helper()
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		_, err := transaction.ExecContext(ctx, `
INSERT INTO work_boards
 (id,creator_user_id,name,description,lifecycle_state,reference_epoch,
  row_version,created_at,updated_at)
VALUES
 ('31000000-0000-4000-8000-000000000010',
  '31000000-0000-4000-8000-000000000001','Query board','query fixture',
  'ACTIVE',1,1,'2026-08-02T08:00:00Z','2026-08-02T08:00:00Z');
INSERT INTO work_board_columns
 (id,board_id,name,status_category,position,row_version,created_at,updated_at,color)
VALUES
 ('31000000-0000-4000-8000-000000000011',
  '31000000-0000-4000-8000-000000000010','Open','OPEN',1,1,
  '2026-08-02T08:00:00Z','2026-08-02T08:00:00Z','#64748b');
INSERT INTO work_board_memberships
 (id,board_id,principal_kind,principal_id,board_role,active,row_version,
  created_at,updated_at)
VALUES
 ('31000000-0000-4000-8000-000000000020',
  '31000000-0000-4000-8000-000000000010','USER',
  '31000000-0000-4000-8000-000000000001','ADMIN',true,1,
  '2026-08-02T08:00:00Z','2026-08-02T08:00:00Z');
INSERT INTO work_tasks
 (id,board_id,column_id,creator_user_id,task_kind,status_category,priority,
  title,description,position,row_version,starts_at,due_at,created_at,updated_at)
VALUES
 ('31000000-0000-4000-8000-000000000012',
  '31000000-0000-4000-8000-000000000010',
  '31000000-0000-4000-8000-000000000011',
  '31000000-0000-4000-8000-000000000001','CUSTOM','OPEN','HIGH',
  'Assigned task','query fixture',1,1,'2026-08-02T09:00:00Z',
  '2026-08-02T10:00:00Z','2026-08-02T08:00:00Z','2026-08-02T08:00:00Z');
INSERT INTO work_task_assignees
 (task_id,principal_kind,principal_id,active,created_at,updated_at)
VALUES
 ('31000000-0000-4000-8000-000000000012','USER',
  '31000000-0000-4000-8000-000000000001',true,
  '2026-08-02T08:00:00Z','2026-08-02T08:00:00Z');
INSERT INTO work_calendar_events
 (id,owner_id,audience,board_id,title,description,color,starts_at,ends_at,
  timezone_name,generated,edited,live_reference_count,legal_hold,row_version,
  created_at,updated_at)
VALUES
 ('31000000-0000-4000-8000-000000000013',
  '31000000-0000-4000-8000-000000000001','BOARD',
  '31000000-0000-4000-8000-000000000010','Board event','query fixture',
  '#64748b','2026-08-02T11:00:00Z','2026-08-02T12:00:00Z','UTC',false,
  false,0,false,1,'2026-08-02T08:00:00Z','2026-08-02T08:00:00Z');
`)
		if err != nil {
			t.Fatal(err)
		}
		seedWorkProjectionRows(t, ctx, transaction)
	})
}

func seedWorkProjectionRows(t *testing.T, ctx context.Context, transaction *sql.Tx) {
	t.Helper()
	_, err := transaction.ExecContext(ctx, `
INSERT INTO work_inbox_items
 (id,source_owner,source_kind,source_id,source_version_id,source_revision,
  permission_revision,event_sequence,summary,due_at,state,visible_to_principals,updated_at)
VALUES
 ('31000000-0000-4000-8000-000000000014','WORKFLOW','HUMAN_TASK',
  '31000000-0000-4000-8000-000000000114','31000000-0000-4000-8000-000000000214',
  1,7,1,'Current inbox','2026-08-02T13:00:00Z','VISIBLE',ARRAY[$1]::text[],
  '2026-08-02T08:00:00Z'),
	 ('31000000-0000-4000-8000-000000000015','WORKFLOW','HUMAN_TASK',
	  '31000000-0000-4000-8000-000000000115','31000000-0000-4000-8000-000000000215',
	  1,6,1,'Stale revision inbox','2026-08-02T14:00:00Z','VISIBLE',ARRAY[$1]::text[],
	  '2026-08-02T08:00:00Z')`, workQueryPrincipal)
	if err != nil {
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx, `INSERT INTO work_calendar_items
	 (id,source_owner,source_kind,source_id,source_version_id,source_revision,
	  permission_revision,event_sequence,summary,starts_at,ends_at,state,
	  visible_to_principals,updated_at)
VALUES
 ('31000000-0000-4000-8000-000000000016','WORKFLOW','HUMAN_TASK',
  '31000000-0000-4000-8000-000000000116','31000000-0000-4000-8000-000000000216',
  1,7,1,'Current calendar','2026-08-02T15:00:00Z','2026-08-02T16:00:00Z',
  'VISIBLE',ARRAY[$1]::text[],'2026-08-02T08:00:00Z'),
	 ('31000000-0000-4000-8000-000000000017','WORKFLOW','HUMAN_TASK',
	  '31000000-0000-4000-8000-000000000117','31000000-0000-4000-8000-000000000217',
	  1,6,1,'Stale revision calendar','2026-08-02T17:00:00Z','2026-08-02T18:00:00Z',
	  'VISIBLE',ARRAY[$1]::text[],'2026-08-02T08:00:00Z')`, workQueryPrincipal)
	if err != nil {
		t.Fatal(err)
	}
}
