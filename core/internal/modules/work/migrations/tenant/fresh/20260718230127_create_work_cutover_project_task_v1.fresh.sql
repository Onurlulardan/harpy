-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION work_v1_adoption_project_task_v1(
    p_target_kind text,
    p_target_id uuid,
    p_record jsonb
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       NOT work_v1_adoption_document_valid_v1(p_target_kind, p_record) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Work V1 task projection denied';
    END IF;
    CASE p_target_kind
    WHEN 'TASK' THEN
        INSERT INTO work_tasks (
            id, board_id, column_id, task_kind, status_category, title,
            description, position, row_version, due_at, completed_at,
            cancelled_at, deleted_at, created_at, updated_at
        ) VALUES (
            p_target_id, (p_record->>'BoardID')::uuid,
            (p_record->>'ColumnID')::uuid, p_record->>'Kind',
            p_record->>'Status', p_record->>'Title', p_record->>'Description',
            (p_record->>'Position')::bigint, 1,
            (p_record->>'DueAt')::timestamptz,
            (p_record->>'CompletedAt')::timestamptz,
            (p_record->>'CancelledAt')::timestamptz,
            (p_record->>'DeletedAt')::timestamptz,
            (p_record->>'CreatedAt')::timestamptz,
            (p_record->>'UpdatedAt')::timestamptz
        )
        ON CONFLICT (id) DO UPDATE SET
            board_id=EXCLUDED.board_id, column_id=EXCLUDED.column_id,
            task_kind=EXCLUDED.task_kind,
            status_category=EXCLUDED.status_category, title=EXCLUDED.title,
            description=EXCLUDED.description, position=EXCLUDED.position,
            due_at=EXCLUDED.due_at, completed_at=EXCLUDED.completed_at,
            cancelled_at=EXCLUDED.cancelled_at, deleted_at=EXCLUDED.deleted_at,
            created_at=EXCLUDED.created_at, updated_at=EXCLUDED.updated_at;
    WHEN 'ASSIGNEE' THEN
        INSERT INTO work_task_assignees (
            task_id, principal_kind, principal_id, active, removed_at,
            created_at, updated_at
        ) VALUES (
            (p_record->>'TaskID')::uuid, p_record->>'PrincipalKind',
            (p_record->>'PrincipalID')::uuid, true, NULL,
            transaction_timestamp(), transaction_timestamp()
        )
        ON CONFLICT (task_id, principal_kind, principal_id) DO UPDATE SET
            active=true, removed_at=NULL, updated_at=EXCLUDED.updated_at;
    WHEN 'LABEL' THEN
        INSERT INTO work_labels (
            id, board_id, name, color, row_version, deleted_at,
            created_at, updated_at
        ) VALUES (
            p_target_id, (p_record->>'BoardID')::uuid, p_record->>'Name',
            p_record->>'Color', 1, NULL,
            transaction_timestamp(), transaction_timestamp()
        )
        ON CONFLICT (id) DO UPDATE SET
            board_id=EXCLUDED.board_id, name=EXCLUDED.name,
            color=EXCLUDED.color, deleted_at=NULL,
            updated_at=EXCLUDED.updated_at;
    WHEN 'TASK_LABEL' THEN
        INSERT INTO work_task_labels (
            task_id, label_id, active, removed_at, created_at, updated_at
        ) VALUES (
            (p_record->>'TaskID')::uuid, (p_record->>'LabelID')::uuid,
            true, NULL, transaction_timestamp(), transaction_timestamp()
        )
        ON CONFLICT (task_id, label_id) DO UPDATE SET
            active=true, removed_at=NULL, updated_at=EXCLUDED.updated_at;
    WHEN 'CHECKLIST' THEN
        INSERT INTO work_checklist_items (
            id, task_id, item_text, completed, position, row_version,
            deleted_at, created_at, updated_at
        ) VALUES (
            p_target_id, (p_record->>'TaskID')::uuid, p_record->>'Text',
            (p_record->>'Completed')::boolean,
            (p_record->>'Position')::bigint, 1, NULL,
            transaction_timestamp(), transaction_timestamp()
        )
        ON CONFLICT (id) DO UPDATE SET
            task_id=EXCLUDED.task_id, item_text=EXCLUDED.item_text,
            completed=EXCLUDED.completed, position=EXCLUDED.position,
            deleted_at=NULL, updated_at=EXCLUDED.updated_at;
    WHEN 'FORM_LINK' THEN
        INSERT INTO work_resource_links (
            id, work_resource_kind, work_resource_id, target_owner,
            target_kind, target_id, target_version_id, usage,
            target_reference_epoch, permission_revision, state, row_version,
            created_at, released_at
        ) VALUES (
            p_target_id, p_record->>'ResourceKind',
            (p_record->>'ResourceID')::uuid, 'FORMS', 'FORM',
            (p_record->>'FormID')::uuid, (p_record->>'VersionID')::uuid,
            'SCHEMA', (p_record->>'ReferenceEpoch')::bigint,
            (p_record->>'PermissionRevision')::bigint,
            'ACTIVE', 1, transaction_timestamp(), NULL
        )
        ON CONFLICT (id) DO UPDATE SET
            target_id=EXCLUDED.target_id,
            target_version_id=EXCLUDED.target_version_id,
            target_reference_epoch=EXCLUDED.target_reference_epoch,
            permission_revision=EXCLUDED.permission_revision,
            state='ACTIVE', released_at=NULL;
    ELSE
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'Work V1 task projection kind is invalid';
    END CASE;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION work_v1_adoption_project_task_v1(
    text, uuid, jsonb
) FROM PUBLIC;

