-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_v1_adoption_bridge_disabled_v1(p_scope_id uuid)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_writer text;
    v_lifecycle text;
    v_generation bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Notifications V1 bridge guard principal denied';
    END IF;
    SELECT authority.writer,authority.lifecycle_state,authority.generation
    INTO STRICT v_writer,v_lifecycle,v_generation
    FROM notifications_writer_authority AS authority
    WHERE authority.singleton FOR UPDATE;
    IF (v_writer='v1' AND v_lifecycle='writable') OR
       EXISTS (SELECT 1 FROM notifications_migration_intents) OR
       EXISTS (SELECT 1 FROM notifications_migration_receipts) OR
       EXISTS (
           SELECT 1 FROM notifications_compatibility_hits AS hit
           WHERE hit.contract_name='notification_intent_bridge_v1'
       ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Notifications V1 bridge is disabled: signed mTLS receipt contract missing';
    END IF;
    RETURN pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        'forgecrud/notifications/v1-bridge-disabled/v1:'||p_scope_id::text||':'||
        v_writer||':'||
        v_lifecycle||':'||v_generation::text||
        ':missing-signed-mtls-stable-effect-receipt-contract','UTF8'
    )),'hex');
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_adoption_bridge_disabled_v1(uuid) FROM PUBLIC;
