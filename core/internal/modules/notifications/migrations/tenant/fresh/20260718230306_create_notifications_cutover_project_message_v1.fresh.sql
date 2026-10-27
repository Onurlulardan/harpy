-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_v1_adoption_project_message_v1(p_record jsonb)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_id uuid;
    v_recipient uuid;
    v_revision bigint;
BEGIN
    v_id := (p_record->>'id')::uuid;
    IF NOT notifications_v1_adoption_target_record_valid_v1(
        'message',p_record->>'id',v_id,p_record
    ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Notifications V1 message projection is invalid';
    END IF;
    v_recipient := (p_record->>'recipient_id')::uuid;
    v_revision := (p_record->>'inbox_revision')::bigint;
    INSERT INTO notifications_messages
    SELECT (pg_catalog.jsonb_populate_record(NULL::notifications_messages,p_record)).*;
    INSERT INTO notifications_inbox_counters(recipient_id,current_revision)
    VALUES (v_recipient,v_revision)
    ON CONFLICT (recipient_id) DO UPDATE
    SET current_revision=GREATEST(
        notifications_inbox_counters.current_revision,EXCLUDED.current_revision
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_adoption_project_message_v1(jsonb) FROM PUBLIC;
