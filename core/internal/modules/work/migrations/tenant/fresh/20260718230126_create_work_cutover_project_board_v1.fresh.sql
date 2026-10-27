-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION work_v1_adoption_project_board_v1(
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
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       NOT work_v1_adoption_document_valid_v1(p_target_kind, p_record) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Work V1 board projection denied';
    END IF;
    CASE p_target_kind
    WHEN 'BOARD' THEN
        INSERT INTO work_boards (
            id, creator_user_id, name, description, lifecycle_state,
            reference_epoch, row_version, created_at, updated_at, archived_at
        ) VALUES (
            p_target_id, (p_record->>'CreatorID')::uuid, p_record->>'Name',
            p_record->>'Description', p_record->>'Lifecycle', 1, 1,
            (p_record->>'CreatedAt')::timestamptz,
            (p_record->>'UpdatedAt')::timestamptz,
            (p_record->>'ArchivedAt')::timestamptz
        )
        ON CONFLICT (id) DO UPDATE SET
            creator_user_id=EXCLUDED.creator_user_id, name=EXCLUDED.name,
            description=EXCLUDED.description,
            lifecycle_state=EXCLUDED.lifecycle_state,
            created_at=EXCLUDED.created_at, updated_at=EXCLUDED.updated_at,
            archived_at=EXCLUDED.archived_at;
    WHEN 'COLUMN' THEN
        INSERT INTO work_board_columns (
            id, board_id, name, status_category, position, row_version,
            created_at, updated_at
        ) VALUES (
            p_target_id, (p_record->>'BoardID')::uuid, p_record->>'Name',
            p_record->>'Status', (p_record->>'Position')::bigint, 1,
            (p_record->>'CreatedAt')::timestamptz,
            (p_record->>'UpdatedAt')::timestamptz
        )
        ON CONFLICT (id) DO UPDATE SET
            board_id=EXCLUDED.board_id, name=EXCLUDED.name,
            status_category=EXCLUDED.status_category, position=EXCLUDED.position,
            created_at=EXCLUDED.created_at, updated_at=EXCLUDED.updated_at;
    WHEN 'MEMBERSHIP' THEN
        INSERT INTO work_board_memberships (
            id, board_id, principal_kind, principal_id, board_role, active,
            row_version, created_at, updated_at, deactivated_at
        ) VALUES (
            p_target_id, (p_record->>'BoardID')::uuid,
            p_record->>'PrincipalKind', (p_record->>'PrincipalID')::uuid,
            p_record->>'Role', true, 1,
            (p_record->>'CreatedAt')::timestamptz,
            (p_record->>'UpdatedAt')::timestamptz, NULL
        )
        ON CONFLICT (id) DO UPDATE SET
            board_id=EXCLUDED.board_id,
            principal_kind=EXCLUDED.principal_kind,
            principal_id=EXCLUDED.principal_id,
            board_role=EXCLUDED.board_role, active=true,
            created_at=EXCLUDED.created_at, updated_at=EXCLUDED.updated_at,
            deactivated_at=NULL;
    ELSE
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'Work V1 board projection kind is invalid';
    END CASE;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION work_v1_adoption_project_board_v1(
    text, uuid, jsonb
) FROM PUBLIC;

