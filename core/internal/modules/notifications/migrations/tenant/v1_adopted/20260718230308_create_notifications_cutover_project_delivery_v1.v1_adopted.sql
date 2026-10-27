-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_v1_adoption_project_delivery_v1(p_record jsonb)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_id uuid;
BEGIN
    v_id := (p_record->>'id')::uuid;
    IF NOT notifications_v1_adoption_target_record_valid_v1(
        'delivery',p_record->>'id',v_id,p_record
    ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Notifications V1 delivery projection is invalid';
    END IF;
    INSERT INTO notifications_deliveries
    SELECT (pg_catalog.jsonb_populate_record(NULL::notifications_deliveries,p_record)).*;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_adoption_project_delivery_v1(jsonb) FROM PUBLIC;
