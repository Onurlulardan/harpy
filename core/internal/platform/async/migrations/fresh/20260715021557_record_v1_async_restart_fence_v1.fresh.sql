-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_record_v1_restart_fence_v1(
    p_run_id uuid,
    p_fence_kind text,
    p_component_name text,
    p_component_contract_version text,
    p_binary_sha256 text,
    p_config_sha256 text,
    p_restart_probe_sha256 text,
    p_fenced_at timestamptz,
    p_restart_verified_at timestamptz,
    p_receipt_sha256 text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run_state text;
    v_captured_at timestamptz;
    v_inserted boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_cluster_admin' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5652',
            MESSAGE = 'V1 async restart fence requires cluster authority';
    END IF;
    SELECT state, captured_at INTO v_run_state, v_captured_at
    FROM platform.async_v1_adoption_runs
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id FOR UPDATE;
    IF NOT FOUND OR v_run_state IS DISTINCT FROM 'disposing' OR
       p_fenced_at >= p_restart_verified_at OR
       p_restart_verified_at > v_captured_at OR
       p_restart_verified_at > pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE = 'P5653',
            MESSAGE = 'V1 async restart fence evidence is invalid';
    END IF;
    INSERT INTO platform.async_v1_restart_fence_receipts (
        run_id, fence_kind, component_name, component_contract_version,
        binary_sha256, config_sha256, restart_probe_sha256,
        fenced_at, restart_verified_at, receipt_sha256
    ) VALUES (
        p_run_id, p_fence_kind, p_component_name, p_component_contract_version,
        p_binary_sha256, p_config_sha256, p_restart_probe_sha256,
        p_fenced_at, p_restart_verified_at, p_receipt_sha256
    ) ON CONFLICT DO NOTHING RETURNING true INTO v_inserted;
    IF NOT COALESCE(v_inserted, false) AND NOT EXISTS (
        SELECT 1 FROM platform.async_v1_restart_fence_receipts
        WHERE run_id OPERATOR(pg_catalog.=) p_run_id
          AND fence_kind OPERATOR(pg_catalog.=) p_fence_kind
          AND component_name OPERATOR(pg_catalog.=) p_component_name
          AND receipt_sha256 OPERATOR(pg_catalog.=) p_receipt_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5654',
            MESSAGE = 'V1 async restart fence replay differs';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_record_v1_restart_fence_v1(
    uuid,text,text,text,text,text,text,timestamptz,timestamptz,text
) FROM PUBLIC;
