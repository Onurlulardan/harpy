-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_v1_adoption_target_record_valid_v1(
    p_target_kind text, p_target_identity text, p_target_id uuid, p_record jsonb
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_relation oid;
    v_identity text;
    v_target_id text;
    v_semantic boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       pg_catalog.jsonb_typeof(p_record) IS DISTINCT FROM 'object' OR
       p_record = '{}'::jsonb OR pg_catalog.octet_length(p_record::text) > 1048576 OR
       p_target_id = '00000000-0000-0000-0000-000000000000'::uuid THEN
        RETURN false;
    END IF;
    v_relation := CASE p_target_kind
        WHEN 'message' THEN 'notifications_messages'::regclass::oid
        WHEN 'effect' THEN 'notifications_effect_receipts'::regclass::oid
        WHEN 'delivery' THEN 'notifications_deliveries'::regclass::oid
        WHEN 'attempt' THEN 'notifications_delivery_attempts'::regclass::oid
        WHEN 'receipt' THEN 'notifications_delivery_receipts'::regclass::oid
        ELSE NULL
    END;
    IF v_relation IS NULL THEN
        RETURN false;
    END IF;
    v_identity := CASE p_target_kind
        WHEN 'message' THEN p_record->>'id'
        WHEN 'effect' THEN p_record->>'message_id'
        WHEN 'delivery' THEN p_record->>'id'
        WHEN 'attempt' THEN p_record->>'id'
        WHEN 'receipt' THEN p_record->>'id'
    END;
    v_target_id := CASE p_target_kind
        WHEN 'effect' THEN p_record->>'message_id'
        ELSE p_record->>'id'
    END;
    v_semantic := CASE p_target_kind
        WHEN 'message' THEN
            p_record->>'producer'='legacy_v1' AND
            p_record->>'template_version_id'=
                '12000000-0000-4000-8000-000000000002' AND
            p_record->>'template_kind'='plain_text' AND p_record->>'locale'='tr' AND
            p_record->>'row_version'='1' AND p_record->>'dismissed_at' IS NULL AND
            p_record->>'purged_at' IS NULL
        WHEN 'effect' THEN
            p_record->>'producer'='legacy_v1' AND p_record->>'outcome'='created' AND
            p_record->>'legacy_receipt_reference'=
                'v1:notifications:'||(p_record->>'message_id')
        WHEN 'delivery' THEN
            p_record->>'provider_config_version_id' IS NULL AND
            p_record->>'channel'='email' AND p_record->>'state'='accepted' AND
            p_record->>'attempt'='1' AND p_record->>'maximum_attempts'='8' AND
            p_record->>'lease_owner'='' AND p_record->>'lease_generation'='1' AND
            p_record->>'lease_until' IS NULL AND p_record->>'outcome_code'='V1_ACCEPTED' AND
            p_record->>'outcome_unknown'='false' AND p_record->>'mandatory'='false' AND
            p_record->>'row_version'='2'
        WHEN 'attempt' THEN
            p_record->>'attempt_number'='1' AND p_record->>'lease_generation'='1' AND
            p_record->>'provider_kind'='smtp' AND
            p_record->>'response_class'='legacy_v1_accepted' AND
            p_record->>'status_code'='' AND p_record->>'error_code'='V1_ACCEPTED' AND
            p_record->>'started_at'=p_record->>'finished_at'
        WHEN 'receipt' THEN
            pg_catalog.octet_length(p_record->>'signature_key_id') BETWEEN 1 AND 256 AND
            p_record->>'signature_key_id' NOT LIKE 'legacy%' AND
            p_record->>'outcome' IN ('delivered','bounced','complained')
        ELSE false
    END;
    RETURN v_semantic AND
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(p_record) AS key)
       IS NOT DISTINCT FROM
       (SELECT pg_catalog.array_agg(attribute.attname::text ORDER BY attribute.attname::text)
        FROM pg_catalog.pg_attribute AS attribute
        WHERE attribute.attrelid=v_relation AND attribute.attnum>0
          AND NOT attribute.attisdropped)
       AND v_identity IS NOT DISTINCT FROM p_target_identity
       AND v_target_id IS NOT DISTINCT FROM p_target_id::text;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_adoption_target_record_valid_v1(
    text,text,uuid,jsonb
) FROM PUBLIC;
