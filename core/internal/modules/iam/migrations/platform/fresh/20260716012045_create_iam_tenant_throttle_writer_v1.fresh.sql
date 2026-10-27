-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_tenant_throttle_record_attempt_v1(
    p_tenant_id uuid, p_purpose text, p_dimension text, p_key_id text, p_bucket_hmac bytea,
    p_group_generation bigint, p_cutover_generation bigint,
    p_policy_version integer, p_window_seconds integer,
    p_block_after integer, p_block_seconds integer
)
RETURNS TABLE (result_attempt_count integer, result_blocked_until timestamptz, result_revision bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_row platform.iam_auth_throttle_buckets%ROWTYPE;
    v_found boolean; v_prior_found boolean;
    v_now timestamptz;
    v_window_start timestamptz; v_window_end timestamptz;
    v_attempt_count integer; v_blocked_until timestamptz;
    v_revision bigint; v_retain_until timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_tenant_id IS NULL OR p_purpose IS NULL OR
       NOT (p_purpose OPERATOR(pg_catalog.=) ANY (ARRAY['login','refresh',
           'password_reset','email_verification','portal_setup','invitation','reauth']::text[])) OR
       p_dimension IS NULL OR NOT (p_dimension OPERATOR(pg_catalog.=) ANY (
           ARRAY['tenant_slug','identifier','ip','identifier_ip','selector']::text[])) OR p_key_id IS NULL OR
       pg_catalog.octet_length(p_key_id) NOT BETWEEN 1 AND 128 OR
       p_key_id OPERATOR(pg_catalog.!~) '^[a-z0-9][a-z0-9._-]{0,127}$' OR
       p_bucket_hmac IS NULL OR pg_catalog.octet_length(p_bucket_hmac) <> 32 OR
       NOT platform.iam_auth_throttle_policy_v1(
           'tenant', p_purpose, p_dimension, p_policy_version,
           p_window_seconds, p_block_after, p_block_seconds
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM tenant throttle input is invalid';
    END IF;
    PERFORM platform.iam_assert_current_tenant_authority_v1(
        p_tenant_id, p_group_generation, p_cutover_generation
    );
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'harpy:iam-throttle:tenant:' || p_tenant_id::text || ':' ||
        COALESCE(p_purpose, '') || ':' || COALESCE(p_dimension, '') || ':' ||
        COALESCE(p_key_id, '') || ':' ||
        COALESCE(pg_catalog.encode(p_bucket_hmac, 'hex'), ''), 0
    ));
    SELECT bucket.* INTO v_row
    FROM platform.iam_auth_throttle_buckets AS bucket
    WHERE bucket.bucket_scope_kind OPERATOR(pg_catalog.=) 'tenant' AND bucket.scope_id OPERATOR(pg_catalog.=) p_tenant_id
      AND bucket.authority_epoch OPERATOR(pg_catalog.=) p_group_generation AND bucket.purpose OPERATOR(pg_catalog.=) p_purpose
      AND bucket.dimension OPERATOR(pg_catalog.=) p_dimension AND bucket.key_id OPERATOR(pg_catalog.=) p_key_id
      AND bucket.bucket_hmac OPERATOR(pg_catalog.=) p_bucket_hmac
    FOR UPDATE;
    v_found := FOUND;
    v_now := pg_catalog.clock_timestamp();
    IF NOT v_found THEN
        SELECT bucket.* INTO v_row
        FROM platform.iam_auth_throttle_buckets AS bucket
        WHERE bucket.bucket_scope_kind OPERATOR(pg_catalog.=) 'tenant' AND bucket.scope_id OPERATOR(pg_catalog.=) p_tenant_id
          AND bucket.authority_epoch OPERATOR(pg_catalog.<) p_group_generation AND bucket.purpose OPERATOR(pg_catalog.=) p_purpose
          AND bucket.dimension OPERATOR(pg_catalog.=) p_dimension AND bucket.key_id OPERATOR(pg_catalog.=) p_key_id
          AND bucket.bucket_hmac OPERATOR(pg_catalog.=) p_bucket_hmac
        ORDER BY bucket.authority_epoch DESC LIMIT 1 FOR SHARE;
        v_prior_found := FOUND;
        IF v_prior_found AND
           v_now OPERATOR(pg_catalog.<=) v_row.last_attempt_at THEN
            v_now := v_row.last_attempt_at OPERATOR(pg_catalog.+)
                INTERVAL '1 microsecond';
        END IF;
        IF v_prior_found AND
           p_policy_version OPERATOR(pg_catalog.<) v_row.policy_version THEN
            RAISE EXCEPTION USING ERRCODE = 'P5722', MESSAGE =
                'IAM tenant throttle prior policy is newer';
        END IF;
        IF v_prior_found AND
           p_policy_version OPERATOR(pg_catalog.=) v_row.policy_version AND
           v_now OPERATOR(pg_catalog.<) v_row.window_end THEN
            IF v_row.attempt_count OPERATOR(pg_catalog.=) 2147483647 THEN
                RAISE EXCEPTION USING ERRCODE = '22003', MESSAGE =
                    'IAM tenant throttle counter overflow';
            END IF;
            v_window_start := v_row.window_start;
            v_window_end := v_row.window_end;
            v_attempt_count := v_row.attempt_count OPERATOR(pg_catalog.+) 1;
            v_blocked_until := v_row.blocked_until;
        ELSE
            v_window_start := v_now;
            v_window_end := v_now OPERATOR(pg_catalog.+)
                pg_catalog.make_interval(secs => p_window_seconds);
            v_attempt_count := 1;
            IF v_prior_found AND v_row.blocked_until OPERATOR(pg_catalog.>) v_now THEN
                v_blocked_until := v_row.blocked_until;
            END IF;
        END IF;
        v_revision := 1;
        IF v_attempt_count OPERATOR(pg_catalog.>=) p_block_after AND (
           v_blocked_until IS NULL OR v_blocked_until OPERATOR(pg_catalog.<) (
               v_now OPERATOR(pg_catalog.+) pg_catalog.make_interval(secs => p_block_seconds))) THEN
            v_blocked_until := v_now OPERATOR(pg_catalog.+)
                pg_catalog.make_interval(secs => p_block_seconds);
        END IF;
    ELSE
        IF v_row.authority_group_generation IS DISTINCT FROM
               p_group_generation OR
           v_row.cutover_generation IS DISTINCT FROM p_cutover_generation OR
           p_policy_version OPERATOR(pg_catalog.<) v_row.policy_version OR
           v_row.revision OPERATOR(pg_catalog.=) 9223372036854775807 THEN
            RAISE EXCEPTION USING ERRCODE = 'P5722', MESSAGE =
                'IAM tenant throttle authority or revision is stale';
        END IF;
        IF v_now OPERATOR(pg_catalog.<=) v_row.last_attempt_at THEN
            v_now := v_row.last_attempt_at OPERATOR(pg_catalog.+)
                INTERVAL '1 microsecond';
        END IF;
        IF v_now OPERATOR(pg_catalog.>=) v_row.window_end OR
           p_policy_version OPERATOR(pg_catalog.>) v_row.policy_version THEN
            v_window_start := v_now;
            v_window_end := v_now OPERATOR(pg_catalog.+)
                pg_catalog.make_interval(secs => p_window_seconds);
            v_attempt_count := 1;
            IF v_row.blocked_until IS NOT NULL AND
               v_row.blocked_until OPERATOR(pg_catalog.>) v_now THEN
                v_blocked_until := v_row.blocked_until;
            END IF;
            IF p_block_after OPERATOR(pg_catalog.=) 1 AND (
               v_blocked_until IS NULL OR v_blocked_until OPERATOR(pg_catalog.<) (
                   v_now OPERATOR(pg_catalog.+)
                   pg_catalog.make_interval(secs => p_block_seconds))) THEN
                v_blocked_until := v_now OPERATOR(pg_catalog.+)
                    pg_catalog.make_interval(secs => p_block_seconds);
            END IF;
        ELSE
            IF v_row.attempt_count OPERATOR(pg_catalog.=) 2147483647 THEN
                RAISE EXCEPTION USING ERRCODE = '22003', MESSAGE =
                    'IAM tenant throttle counter overflow';
            END IF;
            v_window_start := v_row.window_start;
            v_window_end := v_row.window_end;
            v_attempt_count := v_row.attempt_count OPERATOR(pg_catalog.+) 1;
            v_blocked_until := v_row.blocked_until;
            IF v_attempt_count OPERATOR(pg_catalog.>=) p_block_after AND (
               v_blocked_until IS NULL OR v_blocked_until OPERATOR(pg_catalog.<) (
                   v_now OPERATOR(pg_catalog.+)
                   pg_catalog.make_interval(secs => p_block_seconds))) THEN
                v_blocked_until := v_now OPERATOR(pg_catalog.+)
                    pg_catalog.make_interval(secs => p_block_seconds);
            END IF;
        END IF;
        v_revision := v_row.revision OPERATOR(pg_catalog.+) 1;
    END IF;
    v_retain_until := v_window_end OPERATOR(pg_catalog.+) INTERVAL '24 hours';
    IF v_blocked_until IS NOT NULL AND
       v_retain_until OPERATOR(pg_catalog.<) (
           v_blocked_until OPERATOR(pg_catalog.+) INTERVAL '24 hours'
       ) THEN
        v_retain_until := v_blocked_until OPERATOR(pg_catalog.+) INTERVAL '24 hours';
    END IF;
    IF v_found AND v_retain_until OPERATOR(pg_catalog.<) v_row.retain_until THEN
        v_retain_until := v_row.retain_until;
    END IF;
    IF v_found THEN
        UPDATE platform.iam_auth_throttle_buckets AS bucket
        SET policy_version = p_policy_version,
            window_start = v_window_start, window_end = v_window_end,
            attempt_count = v_attempt_count, blocked_until = v_blocked_until,
            last_attempt_at = v_now, revision = v_revision,
            retain_until = v_retain_until
        WHERE bucket.bucket_scope_kind OPERATOR(pg_catalog.=) 'tenant' AND bucket.scope_id OPERATOR(pg_catalog.=) p_tenant_id
          AND bucket.authority_epoch OPERATOR(pg_catalog.=) p_group_generation AND bucket.purpose OPERATOR(pg_catalog.=) p_purpose
          AND bucket.dimension OPERATOR(pg_catalog.=) p_dimension AND bucket.key_id OPERATOR(pg_catalog.=) p_key_id
          AND bucket.bucket_hmac OPERATOR(pg_catalog.=) p_bucket_hmac;
    ELSE
        INSERT INTO platform.iam_auth_throttle_buckets (
            bucket_scope_kind, scope_id, authority_epoch,
            purpose, dimension, key_id, bucket_hmac,
            environment_id, tenant_id, authority_scope_kind, authority_scope_id,
            authority_group_id, authority_group_generation, authority_domain,
            cutover_generation, policy_version, window_start, window_end,
            attempt_count, blocked_until, last_attempt_at, revision, retain_until
        ) VALUES (
            'tenant', p_tenant_id, p_group_generation,
            p_purpose, p_dimension, p_key_id, p_bucket_hmac,
            NULL, p_tenant_id, 'tenant', p_tenant_id, 'iam_system_resources',
            p_group_generation, 'iam', p_cutover_generation, p_policy_version,
            v_window_start, v_window_end, v_attempt_count, v_blocked_until,
            v_now, v_revision, v_retain_until
        );
    END IF;
    result_attempt_count := v_attempt_count;
    result_blocked_until := v_blocked_until;
    result_revision := v_revision;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.iam_tenant_throttle_record_attempt_v1(uuid,text,text,text,bytea,bigint,bigint,integer,integer,integer,integer) FROM PUBLIC;
