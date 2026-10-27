-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION work_v1_adoption_value_valid_v1(
    p_key jsonb,
    p_value jsonb
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_key_fields text;
    v_value_fields text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       pg_catalog.jsonb_typeof(p_key) IS DISTINCT FROM 'object' OR
       pg_catalog.jsonb_typeof(p_value) IS DISTINCT FROM 'object' THEN
        RETURN false;
    END IF;
    SELECT pg_catalog.string_agg(key, ',' ORDER BY key) INTO v_key_fields
    FROM pg_catalog.jsonb_object_keys(p_key) AS key;
    SELECT pg_catalog.string_agg(key, ',' ORDER BY key) INTO v_value_fields
    FROM pg_catalog.jsonb_object_keys(p_value) AS key;
    IF v_key_fields IS DISTINCT FROM
           'source_id,source_kind,target_id,target_kind' OR
       v_value_fields IS DISTINCT FROM
           'source_id,source_kind,source_sha256,target_id,target_kind,target_record' OR
       p_value->>'source_id' IS DISTINCT FROM p_key->>'source_id' OR
       p_value->>'source_kind' IS DISTINCT FROM p_key->>'source_kind' OR
       p_value->>'target_id' IS DISTINCT FROM p_key->>'target_id' OR
       p_value->>'target_kind' IS DISTINCT FROM p_key->>'target_kind' OR
       p_value->>'source_sha256' !~ '^[0-9a-f]{64}$' OR
       p_value->>'source_kind' NOT IN (
           'task_boards','task_board_columns','board_members','board_tasks',
           'board_task_comments','calendar_events'
       ) OR p_value->>'target_kind' NOT IN (
           'BOARD','COLUMN','MEMBERSHIP','TASK','ASSIGNEE','LABEL','TASK_LABEL',
           'CHECKLIST','COMMENT','CALENDAR','RECURRENCE','FORM_LINK',
           'WORKFLOW_PROJECTION'
       ) THEN
        RETURN false;
    END IF;
    BEGIN
        PERFORM (p_value->>'source_id')::uuid, (p_value->>'target_id')::uuid;
    EXCEPTION WHEN invalid_text_representation THEN
        RETURN false;
    END;
    RETURN work_v1_adoption_document_valid_v1(
        p_value->>'target_kind', p_value->'target_record'
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION work_v1_adoption_value_valid_v1(
    jsonb, jsonb
) FROM PUBLIC;
