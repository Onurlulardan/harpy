-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_v1_adoption_project_effect_v1(p_record jsonb)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_id uuid;
BEGIN
    v_id := (p_record->>'message_id')::uuid;
    IF NOT notifications_v1_adoption_target_record_valid_v1(
        'effect',p_record->>'message_id',v_id,p_record
    ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Notifications V1 effect projection is invalid';
    END IF;
    INSERT INTO notifications_effect_receipts
    SELECT (pg_catalog.jsonb_populate_record(
        NULL::notifications_effect_receipts,p_record
    )).*;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_adoption_project_effect_v1(jsonb) FROM PUBLIC;
