-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_purge_data_v2(
    p_analytics_kind text,p_root_id uuid,p_operation_id uuid,
    p_completed_at timestamptz
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_deleted bigint;
BEGIN
    IF p_analytics_kind NOT IN ('endpoint','dashboard') OR
       pg_catalog.uuid_extract_version(p_root_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog purge data input is invalid';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:legal-hold',6100
    ));
    IF EXISTS (
        SELECT 1 FROM analytics_legal_holds AS hold
        WHERE hold.state='active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge is under legal hold';
    END IF;
    IF p_analytics_kind='endpoint' THEN
        PERFORM 1 FROM analytics_saved_queries AS root
        WHERE root.id=p_root_id AND root.lifecycle_state='retired'
          AND root.updated_at<=p_completed_at
        FOR UPDATE;
        IF NOT FOUND OR EXISTS (
            SELECT 1 FROM analytics_saved_query_versions AS version_row
            WHERE version_row.query_id=p_root_id
              AND version_row.updated_at>p_completed_at
        ) THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog purge query differs';
        END IF;
        PERFORM 1 FROM analytics_saved_query_versions AS version_row
        WHERE version_row.query_id=p_root_id
        ORDER BY version_row.id FOR UPDATE;
        UPDATE analytics_saved_query_versions AS version_row
        SET version_state='retired',
            row_revision=version_row.row_revision+1,
            updated_at=p_completed_at
        WHERE version_row.query_id=p_root_id
          AND version_row.version_state IN (
              'published','needs_revalidation'
          );
        UPDATE analytics_saved_queries AS root
        SET current_draft_version_id=NULL,
            row_revision=root.row_revision+1,
            reference_epoch=root.reference_epoch+1,
            updated_at=p_completed_at
        WHERE root.id=p_root_id AND root.lifecycle_state='retired';
    ELSE
        PERFORM 1 FROM analytics_dashboards AS root
        WHERE root.id=p_root_id AND root.lifecycle_state='retired'
          AND root.updated_at<=p_completed_at
        FOR UPDATE;
        IF NOT FOUND OR EXISTS (
            SELECT 1 FROM analytics_dashboard_versions AS version_row
            WHERE version_row.dashboard_id=p_root_id
              AND version_row.updated_at>p_completed_at
        ) THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog purge dashboard differs';
        END IF;
        PERFORM 1 FROM analytics_dashboard_versions AS version_row
        WHERE version_row.dashboard_id=p_root_id
        ORDER BY version_row.id FOR UPDATE;
        UPDATE analytics_dashboard_versions AS version_row
        SET version_state='retired',
            row_revision=version_row.row_revision+1,
            updated_at=p_completed_at
        WHERE version_row.dashboard_id=p_root_id
          AND version_row.version_state='published';
        UPDATE analytics_dashboards AS root
        SET current_draft_version_id=NULL,
            row_revision=root.row_revision+1,
            reference_epoch=root.reference_epoch+1,
            updated_at=p_completed_at
        WHERE root.id=p_root_id AND root.lifecycle_state='retired';
    END IF;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge root changed';
    END IF;
    INSERT INTO analytics_retention_contexts(
        backend_pid,transaction_id,job_id,opened_at
    ) VALUES (
        pg_catalog.pg_backend_pid(),pg_catalog.pg_current_xact_id(),
        p_operation_id,p_completed_at
    );
    IF p_analytics_kind='endpoint' THEN
        DELETE FROM analytics_query_dependencies AS dependency
        USING analytics_saved_query_versions AS version_row
        WHERE version_row.query_id=p_root_id
          AND dependency.query_version_id=version_row.id;
        DELETE FROM analytics_saved_query_versions AS version_row
        WHERE version_row.query_id=p_root_id;
        DELETE FROM analytics_saved_queries AS root
        WHERE root.id=p_root_id AND root.lifecycle_state='retired'
          AND root.current_draft_version_id IS NULL;
    ELSE
        DELETE FROM analytics_dashboard_widgets AS widget
        USING analytics_dashboard_versions AS version_row
        WHERE version_row.dashboard_id=p_root_id
          AND widget.dashboard_version_id=version_row.id;
        DELETE FROM analytics_dashboard_acls AS acl
        WHERE acl.dashboard_id=p_root_id;
        DELETE FROM analytics_dashboard_preferences AS preference
        WHERE preference.dashboard_id=p_root_id;
        DELETE FROM analytics_dashboard_versions AS version_row
        WHERE version_row.dashboard_id=p_root_id;
        DELETE FROM analytics_dashboards AS root
        WHERE root.id=p_root_id AND root.lifecycle_state='retired'
          AND root.current_draft_version_id IS NULL;
    END IF;
    GET DIAGNOSTICS v_deleted=ROW_COUNT;
    IF v_deleted<>1 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge root delete differs';
    END IF;
    DELETE FROM analytics_retention_contexts AS context
    WHERE context.backend_pid=pg_catalog.pg_backend_pid()
      AND context.transaction_id=pg_catalog.pg_current_xact_id()
      AND context.job_id=p_operation_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge context differs';
    END IF;
END;
$function$;
REVOKE ALL ON FUNCTION analytics_catalog_purge_data_v2(
    text,uuid,uuid,timestamptz
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION analytics_catalog_purge_data_v2(
    text,uuid,uuid,timestamptz
) FROM harpy_runtime;
-- +goose StatementEnd
