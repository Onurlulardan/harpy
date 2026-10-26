//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

const workLegacyDDL = `
CREATE TABLE task_boards (
 id uuid PRIMARY KEY, tenant_id uuid NOT NULL, organization_id uuid NOT NULL,
 created_by uuid NOT NULL, name varchar(255) NOT NULL, description text,
 created_at timestamp, updated_at timestamp, deleted_at timestamp
);
CREATE TABLE task_board_columns (
 id uuid PRIMARY KEY, tenant_id uuid NOT NULL, board_id uuid NOT NULL,
 name varchar(255) NOT NULL, color varchar(20), linked_status varchar(30), "order" integer NOT NULL,
 created_at timestamp, updated_at timestamp
);
CREATE TABLE board_members (
 id uuid PRIMARY KEY, tenant_id uuid NOT NULL, board_id uuid NOT NULL,
 user_id uuid NOT NULL, role varchar(20) NOT NULL,
 created_at timestamp, updated_at timestamp
);
CREATE TABLE board_tasks (
 id uuid PRIMARY KEY, tenant_id uuid NOT NULL, board_id uuid NOT NULL,
 column_id uuid NOT NULL, created_by uuid NOT NULL, title varchar(500) NOT NULL,
 description text, task_type varchar(30) NOT NULL, priority varchar(20) NOT NULL,
 status varchar(30) NOT NULL, form_slug varchar(255), assignee_type varchar(20), column_order integer NOT NULL,
 form_data_id uuid, assignee_id uuid, workflow_task_id uuid, due_date timestamp,
 start_date timestamp, completed_at timestamp, recurrence_end timestamp,
 deleted_at timestamp, is_recurring boolean, recurrence_rule varchar(50),
 form_filter jsonb, labels jsonb, checklist jsonb,
 created_at timestamp, updated_at timestamp
);
CREATE TABLE board_task_comments (
 id uuid PRIMARY KEY, tenant_id uuid NOT NULL, task_id uuid NOT NULL,
 user_id uuid NOT NULL, content text NOT NULL, created_at timestamp,
 updated_at timestamp, deleted_at timestamp
);
CREATE TABLE calendar_events (
 id uuid PRIMARY KEY, tenant_id uuid NOT NULL, organization_id uuid NOT NULL,
 created_by uuid NOT NULL, title varchar(255) NOT NULL, description text, color varchar(20),
 start_date timestamp NOT NULL, end_date timestamp NOT NULL, all_day boolean,
 form_slug varchar(255), form_data_id varchar(255), created_at timestamp,
 updated_at timestamp, deleted_at timestamp
);`

type legacyFixture struct {
	tenant, board, column, task, creator, organization uuid.UUID
	name                                               string
	formFilter                                         *string
}

func seedLegacyWork(t *testing.T, ctx context.Context, database *sql.DB,
	schema, name string, seed byte,
) legacyFixture {
	t.Helper()
	fixture := legacyFixture{
		tenant: workUUID(seed, 1), board: workUUID(seed, 2), column: workUUID(seed, 3),
		task: workUUID(seed, 6), creator: workUUID(seed, 4),
		organization: workUUID(seed, 5), name: name,
	}
	var formFilter any
	if seed%2 == 0 {
		value := `{"status":"open"}`
		fixture.formFilter, formFilter = &value, value
	}
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		now := time.Date(2026, 7, 19, 9, 0, 0, 0, time.UTC)
		if _, err := transaction.ExecContext(ctx, `
INSERT INTO task_boards
(id,tenant_id,organization_id,created_by,name,description,created_at,updated_at)
VALUES ($1,$2,$3,$4,$5,'legacy board',$6,$6)`, fixture.board, fixture.tenant,
			fixture.organization, fixture.creator, fixture.name, now); err != nil {
			t.Fatal(err)
		}
		if _, err := transaction.ExecContext(ctx, `
INSERT INTO task_board_columns
(id,tenant_id,board_id,name,color,linked_status,"order",created_at,updated_at)
VALUES ($1,$2,$3,'Open','#64748b','TODO',1,$4,$4)`, fixture.column,
			fixture.tenant, fixture.board, now); err != nil {
			t.Fatal(err)
		}
		if _, err := transaction.ExecContext(ctx, `
INSERT INTO board_tasks
(id,tenant_id,board_id,column_id,created_by,title,description,task_type,priority,
 status,form_slug,assignee_type,column_order,due_date,recurrence_end,is_recurring,
 recurrence_rule,form_filter,labels,checklist,created_at,updated_at)
VALUES ($1,$2,$3,$4,$5,'Legacy task','legacy task','FORM_FILL','MEDIUM','TODO',
 'leave-request',NULL,1,$6::timestamp+interval '1 day',$6::timestamp+interval '10 days',true,
 'FREQ=DAILY;COUNT=3',$7::jsonb,'[]'::jsonb,'[]'::jsonb,$6,$6)`, fixture.task,
			fixture.tenant, fixture.board, fixture.column, fixture.creator, now, formFilter); err != nil {
			t.Fatal(err)
		}
	})
	return fixture
}

func workUUID(seed, suffix byte) uuid.UUID {
	value := uuid.MustParse("019b2ee0-1700-7a1a-8d21-5ca5797f2800")
	value[14], value[15] = seed, suffix
	return value
}
