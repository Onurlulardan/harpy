-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_auth_throttle_blocked_v2(
    p_scope_kind text, p_tenant_id uuid, p_purpose text, p_dimension text,
    p_key_id text, p_bucket_hmac bytea,
    p_group_generation bigint, p_cutover_generation bigint
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_scope_id uuid; v_epoch bigint;
    v_window integer; v_after integer; v_seconds integer;
    v_blocked_until timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_kind NOT IN ('tenant','environment') OR
       p_purpose NOT IN ('login','refresh') OR
       NOT platform.iam_hmac_key_id_valid_v1(p_key_id) OR
       pg_catalog.octet_length(p_bucket_hmac) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM throttle admission input is invalid';
    END IF;
    v_window := 300; v_seconds := 900;
    v_after := CASE p_dimension WHEN 'identifier' THEN 5 WHEN 'selector' THEN 5
        WHEN 'identifier_ip' THEN 10 WHEN 'tenant_slug' THEN 30
        WHEN 'ip' THEN 100 END;
    IF v_after IS NULL OR NOT platform.iam_auth_throttle_policy_v1(
        p_scope_kind,p_purpose,p_dimension,1,v_window,v_after,v_seconds
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM throttle admission policy is invalid';
    END IF;
    IF p_scope_kind = 'tenant' THEN
        IF p_tenant_id IS NULL OR p_group_generation IS NULL OR
           p_cutover_generation IS NULL THEN
            RAISE EXCEPTION USING ERRCODE = '42501',
                MESSAGE = 'IAM tenant throttle authority is invalid';
        END IF;
        PERFORM platform.iam_assert_current_tenant_authority_v1(
            p_tenant_id,p_group_generation,p_cutover_generation
        );
        v_scope_id := p_tenant_id; v_epoch := p_group_generation;
    ELSE
        IF p_tenant_id IS NOT NULL OR p_group_generation IS NOT NULL OR
           p_cutover_generation IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '42501',
                MESSAGE = 'IAM environment throttle authority is invalid';
        END IF;
        SELECT environment_id INTO STRICT v_scope_id
        FROM platform.environment_identity FOR SHARE;
        v_epoch := 0;
    END IF;
    SELECT bucket.blocked_until INTO v_blocked_until
    FROM platform.iam_auth_throttle_buckets AS bucket
    WHERE bucket.bucket_scope_kind = p_scope_kind
      AND bucket.scope_id = v_scope_id AND bucket.purpose = p_purpose
      AND bucket.dimension = p_dimension AND bucket.key_id = p_key_id
      AND bucket.bucket_hmac = p_bucket_hmac
      AND (p_scope_kind = 'environment' OR bucket.authority_epoch <= v_epoch)
    ORDER BY bucket.authority_epoch DESC LIMIT 1;
    RETURN v_blocked_until IS NOT NULL AND
        v_blocked_until > pg_catalog.clock_timestamp();
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_auth_throttle_blocked_v2(
    text,uuid,text,text,text,bytea,bigint,bigint
) FROM PUBLIC;
