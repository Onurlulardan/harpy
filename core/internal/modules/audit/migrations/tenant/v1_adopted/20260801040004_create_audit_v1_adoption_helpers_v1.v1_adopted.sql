-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION audit_v1_owner_module_v1(p_entity_type text) RETURNS text
LANGUAGE sql IMMUTABLE SET search_path = pg_catalog AS $$
SELECT CASE pg_catalog.lower(p_entity_type)
    WHEN 'form_record' THEN 'forms' WHEN 'form_config' THEN 'forms'
    WHEN 'user' THEN 'iam' WHEN 'role' THEN 'iam'
    WHEN 'organization' THEN 'iam' WHEN 'permission' THEN 'iam'
    WHEN 'document' THEN 'documents' WHEN 'workflow' THEN 'workflow'
    ELSE 'audit' END;
$$;

CREATE FUNCTION audit_v1_operation_v1(p_entity_type text, p_action text) RETURNS text
LANGUAGE plpgsql IMMUTABLE SET search_path = pg_catalog AS $$
DECLARE v_prefix text;
BEGIN
    IF pg_catalog.lower(p_action) NOT IN ('create','update','delete') THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='AUDIT_V1_ACTION_INVALID';
    END IF;
    v_prefix := CASE pg_catalog.lower(p_entity_type)
        WHEN 'form_record' THEN 'forms.record' WHEN 'form_config' THEN 'forms.definition'
        WHEN 'user' THEN 'iam.user' WHEN 'role' THEN 'iam.role'
        WHEN 'organization' THEN 'iam.organization' WHEN 'permission' THEN 'iam.permission'
        WHEN 'document' THEN 'documents.item' WHEN 'workflow' THEN 'workflow.definition'
        ELSE 'audit.legacy' END;
    RETURN v_prefix || '.' || pg_catalog.lower(p_action);
END;
$$;

CREATE FUNCTION audit_v1_run_document_v1(p_run audit_v1_adoption_runs) RETURNS jsonb
LANGUAGE sql STABLE SET search_path FROM CURRENT AS $$
SELECT pg_catalog.jsonb_build_object(
    'run_id',p_run.run_id,'status',p_run.status,'source_count',p_run.source_count,
    'adopted_count',p_run.adopted_count,'next_page',p_run.next_page,
    'cursor_created_at',p_run.cursor_created_at,'cursor_id',p_run.cursor_id,
    'started_at',p_run.started_at,'completed_at',p_run.completed_at);
$$;
-- +goose StatementEnd
