-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_retention_purge_audits_v1(
    p_execution_cutoff timestamptz,p_audit_cutoff timestamptz,
    p_capability_cutoff timestamptz,p_limit integer
)
RETURNS TABLE (executions_purged bigint,audits_purged bigint,capabilities_purged bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_remaining integer;
    v_approval_count bigint;
    v_operation_count bigint;
BEGIN
    IF p_execution_cutoff IS NULL OR NOT pg_catalog.isfinite(p_execution_cutoff) OR
       p_audit_cutoff IS NULL OR NOT pg_catalog.isfinite(p_audit_cutoff) OR
       p_capability_cutoff IS NULL OR NOT pg_catalog.isfinite(p_capability_cutoff) OR
       p_limit NOT BETWEEN 1 AND 100 OR
       NOT analytics_retention_delete_allowed_v1(pg_catalog.pg_backend_pid()) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics audit retention authority is invalid';
    END IF;
    WITH candidates AS MATERIALIZED (
        SELECT execution.execution_id
        FROM analytics_execution_audits AS execution
        WHERE execution.occurred_at <= p_execution_cutoff
        ORDER BY execution.occurred_at,execution.execution_id
        FOR UPDATE SKIP LOCKED LIMIT p_limit
    )
    DELETE FROM analytics_execution_audits AS execution
    USING candidates WHERE execution.execution_id = candidates.execution_id;
    GET DIAGNOSTICS executions_purged = ROW_COUNT;
    v_remaining := p_limit - executions_purged::integer;
    IF v_remaining < 1 THEN
        audits_purged := 0; capabilities_purged := 0;
        RETURN NEXT; RETURN;
    END IF;
    WITH candidates AS MATERIALIZED (
        SELECT approval.release_id
        FROM analytics_publication_approvals AS approval
        WHERE approval.approved_at <= p_audit_cutoff
        ORDER BY approval.approved_at,approval.release_id
        FOR UPDATE SKIP LOCKED LIMIT v_remaining
    )
    DELETE FROM analytics_publication_approvals AS approval
    USING candidates WHERE approval.release_id = candidates.release_id;
    GET DIAGNOSTICS v_approval_count = ROW_COUNT;
    v_remaining := v_remaining - v_approval_count::integer;
    v_operation_count := 0;
    IF v_remaining > 0 THEN
        WITH candidates AS MATERIALIZED (
            SELECT operation.operation_id
            FROM analytics_operations AS operation
            WHERE operation.occurred_at <= p_audit_cutoff
            ORDER BY operation.occurred_at,operation.operation_id
            FOR UPDATE SKIP LOCKED LIMIT v_remaining
        )
        DELETE FROM analytics_operations AS operation
        USING candidates WHERE operation.operation_id = candidates.operation_id;
        GET DIAGNOSTICS v_operation_count = ROW_COUNT;
        v_remaining := v_remaining - v_operation_count::integer;
    END IF;
    audits_purged := v_approval_count + v_operation_count;
    capabilities_purged := 0;
    IF v_remaining > 0 THEN
        WITH candidates AS MATERIALIZED (
            SELECT token.release_id
            FROM analytics_capability_tokens AS token
            WHERE token.revoked_at IS NOT NULL
              AND token.expires_at <= p_capability_cutoff
            ORDER BY token.expires_at,token.release_id
            FOR UPDATE SKIP LOCKED LIMIT v_remaining
        )
        DELETE FROM analytics_capability_tokens AS token
        USING candidates WHERE token.release_id = candidates.release_id;
        GET DIAGNOSTICS capabilities_purged = ROW_COUNT;
    END IF;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_retention_purge_audits_v1(
    timestamptz,timestamptz,timestamptz,integer
) FROM PUBLIC;
