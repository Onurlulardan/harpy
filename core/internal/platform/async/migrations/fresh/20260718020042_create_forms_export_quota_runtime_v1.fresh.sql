-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_export_quota_release_v1(
    p_tenant_id uuid,
    p_mode text,
    p_ciphertext_bytes bigint,
    p_observed_at timestamptz
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_revision bigint;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_mode NOT IN ('unknown_reserved','expired_used') OR
       NOT pg_catalog.isfinite(p_observed_at) OR
       (p_mode='unknown_reserved' AND p_ciphertext_bytes<>2147483648) OR
       (p_mode='expired_used' AND
        p_ciphertext_bytes NOT BETWEEN 1 AND 2147483648) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'Forms export quota release input is invalid';
    END IF;
    UPDATE platform.artifact_quota_counters
    SET reserved_ciphertext_bytes=reserved_ciphertext_bytes-
            CASE WHEN p_mode='unknown_reserved' THEN p_ciphertext_bytes ELSE 0 END,
        used_ciphertext_bytes=used_ciphertext_bytes-
            CASE WHEN p_mode='expired_used' THEN p_ciphertext_bytes ELSE 0 END,
        revision=revision+1,updated_at=p_observed_at
    WHERE tenant_id=p_tenant_id
      AND (p_mode<>'unknown_reserved' OR
           reserved_ciphertext_bytes>=p_ciphertext_bytes)
      AND (p_mode<>'expired_used' OR used_ciphertext_bytes>=p_ciphertext_bytes)
    RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P6049',
            MESSAGE = 'Forms export quota release is inconsistent';
    END IF;
    RETURN v_revision;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_forms_export_quota_release_v1(
    uuid,text,bigint,timestamptz
) FROM PUBLIC;

CREATE FUNCTION platform.platform_async_forms_export_quota_state_v1(
    p_tenant_id uuid
)
RETURNS TABLE(
    used_ciphertext_bytes bigint,
    reserved_ciphertext_bytes bigint,
    fenced boolean,
    fence_code text,
    revision bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms export quota read authority is invalid';
    END IF;
    RETURN QUERY SELECT quota.used_ciphertext_bytes,
        quota.reserved_ciphertext_bytes,quota.fenced,quota.fence_code,quota.revision
    FROM platform.artifact_quota_counters AS quota
    WHERE quota.tenant_id=p_tenant_id;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_forms_export_quota_state_v1(uuid)
    FROM PUBLIC;

CREATE FUNCTION platform.platform_async_forms_export_quota_fence_v1(
    p_tenant_id uuid,
    p_expected_revision bigint,
    p_fenced boolean,
    p_fence_code text,
    p_observed_at timestamptz
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE v_revision bigint;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_expected_revision<0 OR p_fenced<>(p_fence_code IS NOT NULL) OR
       (p_fence_code IS NOT NULL AND
        p_fence_code !~ '^[A-Z][A-Z0-9_]{1,63}$') OR
       NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'Forms export quota fence input is invalid';
    END IF;
    UPDATE platform.artifact_quota_counters SET fenced=p_fenced,
        fence_code=p_fence_code,revision=revision+1,updated_at=p_observed_at
    WHERE tenant_id=p_tenant_id AND revision=p_expected_revision
    RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'Forms export quota fence revision changed';
    END IF;
    RETURN v_revision;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_forms_export_quota_fence_v1(
    uuid,bigint,boolean,text,timestamptz
) FROM PUBLIC;
-- +goose StatementEnd
