-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.support_require_adopter_v1()
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $$
BEGIN
    IF session_user <> 'forgecrud_adopter' THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='support adopter authority required';
    END IF;
END;
$$;

CREATE FUNCTION platform.support_adoption_run_document_v1(
    p_run platform.support_adoption_runs
)
RETURNS jsonb
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog
AS $$
    SELECT pg_catalog.jsonb_build_object(
        'operation_id', p_run.operation_id,
        'run_id', p_run.run_id,
        'source_kind', p_run.source_kind,
        'source_tenant_id', p_run.source_tenant_id,
        'source_schema', p_run.source_schema,
        'source_relation', p_run.source_relation,
        'source_relation_oid', p_run.source_relation_oid::bigint,
        'tenant_slug', p_run.tenant_slug,
        'tenant_name', p_run.tenant_name,
        'snapshot_id', p_run.snapshot_id,
        'source_locator_sha256', pg_catalog.encode(p_run.source_locator_sha256,'hex'),
        'source_manifest_sha256', pg_catalog.encode(p_run.source_manifest_sha256,'hex'),
        'transform_sha256', pg_catalog.encode(p_run.transform_sha256,'hex'),
        'expected_count', p_run.expected_count,
        'copied_count', p_run.copied_count,
        'cursor_id', p_run.cursor_id,
        'next_page', p_run.next_page,
        'expected_source_sha256', pg_catalog.encode(p_run.expected_source_sha256,'hex'),
        'source_sha256', pg_catalog.encode(p_run.source_sha256,'hex'),
        'target_sha256', pg_catalog.encode(p_run.target_sha256,'hex'),
        'status', p_run.status,
        'revision', p_run.revision,
        'lease_owner', p_run.lease_owner,
        'lease_generation', p_run.lease_generation,
        'lease_expires_at', p_run.lease_expires_at,
        'created_at', p_run.created_at,
        'updated_at', p_run.updated_at
    )
$$;

CREATE FUNCTION platform.support_adoption_page_document_v1(
    p_page platform.support_adoption_page_receipts
)
RETURNS jsonb
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog
AS $$
    SELECT pg_catalog.jsonb_build_object(
        'run_id', p_page.run_id,
        'page_number', p_page.page_number,
        'cursor_before', p_page.cursor_before,
        'cursor_after', p_page.cursor_after,
        'row_count', p_page.row_count,
        'source_page_sha256', pg_catalog.encode(p_page.source_page_sha256,'hex'),
        'target_page_sha256', pg_catalog.encode(p_page.target_page_sha256,'hex'),
        'advanced', p_page.state = 'advanced'
    )
$$;

CREATE FUNCTION platform.support_contact_adoption_document_v1(p_id uuid)
RETURNS jsonb
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog
AS $$
    SELECT pg_catalog.jsonb_build_object(
        'id', message.id, 'name', message.name, 'email', message.email,
        'company', message.company, 'phone', message.phone,
        'employee_count', message.employee_count, 'message', message.message,
        'contact_type', message.contact_type, 'status', message.status,
        'source', message.source, 'notes', message.notes,
        'replied_at', message.replied_at, 'replied_by', message.replied_by,
        'revision', message.revision, 'created_at', message.created_at,
        'updated_at', message.updated_at)
    FROM platform.support_contact_messages AS message WHERE message.id = p_id
$$;

CREATE FUNCTION platform.support_bug_adoption_document_v1(p_id uuid)
RETURNS jsonb
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog
AS $$
    SELECT pg_catalog.jsonb_build_object(
        'id', report.id, 'tenant_id', report.tenant_id,
        'tenant_slug', report.tenant_slug, 'tenant_name', report.tenant_name,
        'reporter_id', report.reporter_id, 'reporter_name', report.reporter_name,
        'reporter_email', report.reporter_email, 'title', report.title,
        'description', report.description, 'category', report.category,
        'severity', report.severity, 'priority', report.priority,
        'page_url', report.page_url, 'user_agent', report.user_agent,
        'browser_info', report.browser_info, 'console_errors', report.console_errors,
        'network_info', report.network_info, 'screenshots', report.legacy_screenshot_refs,
        'steps_to_reproduce', report.steps_to_reproduce, 'status', report.status,
        'resolution', report.resolution, 'admin_notes', report.admin_notes,
        'resolved_at', report.resolved_at, 'closed_at', report.closed_at,
        'revision', report.revision, 'created_at', report.created_at,
        'updated_at', report.updated_at)
    FROM platform.support_bug_reports AS report WHERE report.id = p_id
$$;

REVOKE ALL ON FUNCTION platform.support_require_adopter_v1() FROM PUBLIC, forgecrud_runtime, forgecrud_adopter;
REVOKE ALL ON FUNCTION platform.support_adoption_run_document_v1(platform.support_adoption_runs) FROM PUBLIC, forgecrud_runtime, forgecrud_adopter;
REVOKE ALL ON FUNCTION platform.support_adoption_page_document_v1(platform.support_adoption_page_receipts) FROM PUBLIC, forgecrud_runtime, forgecrud_adopter;
REVOKE ALL ON FUNCTION platform.support_contact_adoption_document_v1(uuid) FROM PUBLIC, forgecrud_runtime, forgecrud_adopter;
REVOKE ALL ON FUNCTION platform.support_bug_adoption_document_v1(uuid) FROM PUBLIC, forgecrud_runtime, forgecrud_adopter;
-- +goose StatementEnd

