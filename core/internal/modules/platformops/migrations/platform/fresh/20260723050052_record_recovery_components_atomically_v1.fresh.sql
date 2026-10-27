-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_record_recovery_components_v1(
    p_recovery_set_id uuid,
    p_attempt_id uuid,
    p_lease_token uuid,
    p_components jsonb
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_kind text;
    v_value jsonb;
    v_changed boolean;
BEGIN
    IF pg_catalog.jsonb_typeof(p_components)<>'object'
       OR pg_catalog.jsonb_object_length(p_components)<>5
       OR NOT p_components ?& ARRAY[
           'database_archive','documents_inventory','key_custody',
           'owner_anchor','protected_authority'
       ] THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery component batch denied';
    END IF;
    FOREACH v_kind IN ARRAY ARRAY[
        'database_archive','documents_inventory','key_custody',
        'owner_anchor','protected_authority'
    ] LOOP
        v_value := p_components->v_kind;
        IF pg_catalog.jsonb_typeof(v_value)<>'object'
           OR pg_catalog.jsonb_object_length(v_value)<>4
           OR (v_value->>'ciphertext_sha256') !~ '^[0-9a-f]{64}$'
           OR (v_value->>'ciphertext_bytes') !~ '^[1-9][0-9]*$' THEN
            RAISE EXCEPTION USING ERRCODE='42501',
                MESSAGE='recovery component batch denied';
        END IF;
        SELECT platform.platformops_record_recovery_component_v3(
            p_recovery_set_id,p_attempt_id,p_lease_token,v_kind,
            v_value->>'object_key',v_value->>'object_version_id',
            v_value->>'ciphertext_sha256',
            (v_value->>'ciphertext_bytes')::bigint
        ) INTO v_changed;
        IF NOT v_changed THEN RETURN false; END IF;
    END LOOP;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_record_recovery_components_v1(
    uuid,uuid,uuid,jsonb
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.platformops_record_recovery_component_v3(
    uuid,uuid,uuid,text,text,text,text,bigint
) FROM forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_record_recovery_components_v1(
    uuid,uuid,uuid,jsonb
) TO forgecrud_backup_reader;
