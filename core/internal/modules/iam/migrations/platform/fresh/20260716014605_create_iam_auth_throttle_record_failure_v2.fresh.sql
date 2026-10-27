-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_auth_throttle_record_failure_v2(
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
    v_scope_id uuid;
    v_epoch bigint;
    v_row platform.iam_auth_throttle_buckets%ROWTYPE;
    v_found boolean;
    v_now timestamptz;
    v_window integer;
    v_after integer;
    v_seconds integer;
    v_start timestamptz;
    v_end timestamptz;
    v_count integer;
    v_blocked timestamptz;
    v_revision bigint;
    v_retain timestamptz;
BEGIN
	 v_window := 300;
	 v_seconds := 900;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_kind NOT IN ('tenant','environment') OR
       p_purpose NOT IN ('login','refresh') OR
       NOT platform.iam_hmac_key_id_valid_v1(p_key_id) OR
       pg_catalog.octet_length(p_bucket_hmac) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'IAM throttle failure input is invalid';
    END IF;
    v_after := CASE p_dimension WHEN 'identifier' THEN 5 WHEN 'selector' THEN 5
        WHEN 'identifier_ip' THEN 10 WHEN 'tenant_slug' THEN 30
        WHEN 'ip' THEN 100 END;
    IF v_after IS NULL OR NOT platform.iam_auth_throttle_policy_v1(
        p_scope_kind,p_purpose,p_dimension,1,v_window,v_after,v_seconds
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'IAM throttle failure policy is invalid';
    END IF;
    IF p_scope_kind = 'tenant' THEN
        IF p_tenant_id IS NULL OR p_group_generation IS NULL OR p_cutover_generation IS NULL THEN
            RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'IAM tenant throttle authority is invalid';
        END IF;
        PERFORM platform.iam_assert_current_tenant_authority_v1(
            p_tenant_id,p_group_generation,p_cutover_generation
        );
        v_scope_id := p_tenant_id; v_epoch := p_group_generation;
    ELSE
        IF p_tenant_id IS NOT NULL OR p_group_generation IS NOT NULL OR p_cutover_generation IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'IAM environment throttle authority is invalid';
        END IF;
        SELECT environment_id INTO STRICT v_scope_id FROM platform.environment_identity FOR SHARE;
        v_epoch := 0;
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'harpy:iam-throttle:v2:' || p_scope_kind || ':' || v_scope_id::text || ':' ||
        p_purpose || ':' || p_dimension || ':' || p_key_id || ':' ||
        pg_catalog.encode(p_bucket_hmac,'hex'),0
    ));
    SELECT bucket.* INTO v_row FROM platform.iam_auth_throttle_buckets AS bucket
    WHERE bucket.bucket_scope_kind = p_scope_kind AND bucket.scope_id = v_scope_id
      AND bucket.authority_epoch = v_epoch AND bucket.purpose = p_purpose
      AND bucket.dimension = p_dimension AND bucket.key_id = p_key_id
      AND bucket.bucket_hmac = p_bucket_hmac FOR UPDATE;
    v_found := FOUND; v_now := pg_catalog.clock_timestamp();
    IF NOT v_found AND p_scope_kind = 'tenant' THEN
        SELECT bucket.* INTO v_row FROM platform.iam_auth_throttle_buckets AS bucket
        WHERE bucket.bucket_scope_kind = 'tenant' AND bucket.scope_id = v_scope_id
          AND bucket.authority_epoch < v_epoch AND bucket.purpose = p_purpose
          AND bucket.dimension = p_dimension AND bucket.key_id = p_key_id
          AND bucket.bucket_hmac = p_bucket_hmac
        ORDER BY bucket.authority_epoch DESC LIMIT 1 FOR SHARE;
    END IF;
    IF FOUND AND v_now <= v_row.last_attempt_at THEN
        v_now := v_row.last_attempt_at + INTERVAL '1 microsecond';
    END IF;
    IF FOUND AND v_now < v_row.window_end AND v_row.policy_version = 1 THEN
        IF v_row.attempt_count = 2147483647 THEN
            RAISE EXCEPTION USING ERRCODE = '22003', MESSAGE = 'IAM throttle counter overflow';
        END IF;
        v_start := v_row.window_start; v_end := v_row.window_end;
        v_count := v_row.attempt_count + 1; v_blocked := v_row.blocked_until;
    ELSE
        v_start := v_now; v_end := v_now + pg_catalog.make_interval(secs => v_window);
        v_count := 1;
        IF FOUND AND v_row.blocked_until > v_now THEN v_blocked := v_row.blocked_until; END IF;
    END IF;
    IF v_count >= v_after AND (v_blocked IS NULL OR
       v_blocked < v_now + pg_catalog.make_interval(secs => v_seconds)) THEN
        v_blocked := v_now + pg_catalog.make_interval(secs => v_seconds);
    END IF;
    v_revision := CASE WHEN v_found THEN v_row.revision + 1 ELSE 1 END;
    v_retain := v_end + INTERVAL '24 hours';
    IF v_blocked IS NOT NULL THEN v_retain := GREATEST(v_retain,v_blocked + INTERVAL '24 hours'); END IF;
    IF v_found THEN v_retain := GREATEST(v_retain,v_row.retain_until); END IF;
    IF v_found THEN
        UPDATE platform.iam_auth_throttle_buckets SET policy_version=1,
            window_start=v_start,window_end=v_end,attempt_count=v_count,
            blocked_until=v_blocked,last_attempt_at=v_now,revision=v_revision,
            retain_until=v_retain
        WHERE bucket_scope_kind=p_scope_kind AND scope_id=v_scope_id
          AND authority_epoch=v_epoch AND purpose=p_purpose AND dimension=p_dimension
          AND key_id=p_key_id AND bucket_hmac=p_bucket_hmac;
    ELSE
        INSERT INTO platform.iam_auth_throttle_buckets(
            bucket_scope_kind,scope_id,authority_epoch,purpose,dimension,key_id,bucket_hmac,
            environment_id,tenant_id,authority_scope_kind,authority_scope_id,
            authority_group_id,authority_group_generation,authority_domain,cutover_generation,
            policy_version,window_start,window_end,attempt_count,blocked_until,
            last_attempt_at,revision,retain_until
        ) VALUES (
            p_scope_kind,v_scope_id,v_epoch,p_purpose,p_dimension,p_key_id,p_bucket_hmac,
            CASE WHEN p_scope_kind='environment' THEN v_scope_id END,p_tenant_id,
            CASE WHEN p_scope_kind='tenant' THEN 'tenant' END,p_tenant_id,
            CASE WHEN p_scope_kind='tenant' THEN 'iam_system_resources' END,p_group_generation,
            CASE WHEN p_scope_kind='tenant' THEN 'iam' END,p_cutover_generation,
            1,v_start,v_end,v_count,v_blocked,v_now,1,v_retain
        );
    END IF;
    RETURN v_blocked IS NOT NULL AND v_blocked > v_now;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_auth_throttle_record_failure_v2(
    text,uuid,text,text,text,bytea,bigint,bigint
) FROM PUBLIC;
