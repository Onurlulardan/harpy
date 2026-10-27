-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_protected_authority_head_initialize_v1(
    p_environment_id uuid, p_registry_binding_id uuid,
    p_generation bigint, p_state_sha256 bytea, p_ack_sha256 bytea
)
RETURNS TABLE (generation bigint, state_sha256 bytea, replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_environment_id IS NULL OR p_registry_binding_id IS NULL OR p_generation < 1 OR
       pg_catalog.octet_length(p_state_sha256) <> 32 OR
       pg_catalog.octet_length(p_ack_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Protected authority head is invalid';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.iam_protected_authority_heads AS head
               WHERE head.environment_id = p_environment_id) THEN
        IF NOT EXISTS (SELECT 1 FROM platform.iam_protected_authority_heads AS head
            WHERE head.environment_id = p_environment_id
              AND head.registry_binding_id = p_registry_binding_id
              AND head.scope_kind = 'environment' AND head.scope_id = p_environment_id
              AND head.generation = p_generation AND head.state_sha256 = p_state_sha256
              AND head.ack_sha256 = p_ack_sha256) THEN
            RAISE EXCEPTION USING ERRCODE = 'P1401', MESSAGE = 'Protected authority head conflicts';
        END IF;
        RETURN QUERY SELECT p_generation, p_state_sha256, true;
        RETURN;
    END IF;
    INSERT INTO platform.iam_protected_authority_heads (
        environment_id, registry_binding_id, scope_kind, scope_id, generation,
        state_sha256, ack_sha256, revision, created_at, updated_at
    ) VALUES (p_environment_id, p_registry_binding_id, 'environment', p_environment_id,
        p_generation, p_state_sha256, p_ack_sha256, 1, v_now, v_now);
    RETURN QUERY SELECT p_generation, p_state_sha256, false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.iam_protected_authority_head_initialize_v1(
    uuid,uuid,bigint,bytea,bytea
) FROM PUBLIC;
