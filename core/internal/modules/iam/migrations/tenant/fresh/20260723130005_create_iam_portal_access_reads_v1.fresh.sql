-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_portal_user_read_v1(
    p_scope_schema name,p_user_id uuid
)
RETURNS TABLE (
    result_id uuid,binding_id uuid,email text,status text,account_type text,
    subject_type text,subject_record_id uuid,result_revision bigint,
    setup_required boolean,created_at timestamptz,updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR p_user_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501';
    END IF;
    RETURN QUERY
    SELECT subject.id,binding.id,email.normalized_value,
      CASE
        WHEN subject.status<>'active' THEN 'INACTIVE'
        WHEN binding.state='active' THEN 'ACTIVE'
        WHEN binding.state='pending' THEN 'PENDING'
        WHEN setup.approval_state IN('pending','revision_required') THEN 'PENDING'
        ELSE 'INACTIVE'
      END,
      CASE WHEN console.id IS NULL THEN 'portal' ELSE 'operator_portal' END,
      binding.subject_type,binding.subject_record_id,subject.revision,
      setup.setup_state='required',subject.created_at,subject.updated_at
    FROM iam_users subject
    JOIN iam_user_identifiers email ON email.user_id=subject.id
      AND email.kind='email_v1' AND email.state<>'retired'
    JOIN iam_portal_subject_bindings binding ON binding.user_id=subject.id
      AND binding.state<>'retired'
    JOIN iam_portal_access_setups setup ON setup.binding_id=binding.id
    LEFT JOIN iam_user_access_modes console ON console.user_id=subject.id
      AND console.mode='operator_console' AND console.state='active'
    WHERE subject.id=p_user_id AND subject.status<>'tombstoned';
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='P0002'; END IF;
END;
$function$;

CREATE FUNCTION iam_portal_access_verify_v1(
    p_scope_schema name,p_tenant_id uuid,p_principal_id uuid,p_binding_id uuid,
    p_token_version bigint,p_authorization_version bigint
)
RETURNS TABLE (
    principal_id uuid,binding_id uuid,tenant_id uuid,token_version bigint,
    authorization_version bigint,authorization_clock bigint,
    binding_revision bigint,subject_type text,form_id uuid,
    form_version_id uuid,subject_record_id uuid,identifier_field_id uuid,
    role_names text[]
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS NULL OR p_principal_id IS NULL OR p_binding_id IS NULL OR
       p_token_version<1 OR p_authorization_version<1 OR
       NOT EXISTS(SELECT 1 FROM tenancy_schema_identity identity_row
          WHERE identity_row.singleton AND identity_row.tenant_id=p_tenant_id
            AND identity_row.schema_name=p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE='42501';
    END IF;
    RETURN QUERY
    SELECT subject.id,binding.id,p_tenant_id,subject.token_version,
      subject.authorization_version,clock.clock_value,binding.revision,
      binding.subject_type,binding.profile_form_id,binding.profile_form_version_id,
      binding.subject_record_id,binding.identifier_field_id,
      (SELECT pg_catalog.array_agg(role.name ORDER BY role.name)
       FROM iam_portal_role_assignments portal_role
       JOIN iam_roles role ON role.id=portal_role.role_id AND role.status='active'
       JOIN iam_user_role_assignments assignment
         ON assignment.id=portal_role.assignment_id AND assignment.state='active'
       WHERE portal_role.binding_id=binding.id)
    FROM iam_users subject
    JOIN iam_portal_subject_bindings binding ON binding.user_id=subject.id
      AND binding.id=p_binding_id AND binding.state='active'
    JOIN iam_user_access_modes access ON access.id=binding.access_mode_id
      AND access.user_id=subject.id AND access.mode='portal' AND access.state='active'
    JOIN iam_portal_access_setups setup ON setup.binding_id=binding.id
      AND setup.setup_state='completed'
      AND setup.approval_state IN('not_required','approved')
    JOIN iam_user_identifiers identifier ON identifier.user_id=subject.id
      AND identifier.kind='portal_v1' AND identifier.state='active'
    CROSS JOIN iam_authorization_clock clock
    WHERE subject.id=p_principal_id AND subject.status='active'
      AND subject.token_version=p_token_version
      AND subject.authorization_version=p_authorization_version
      AND clock.singleton AND clock.clock_value>0;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='42501',
      MESSAGE='IAM portal access is not current'; END IF;
END;
$function$;

CREATE FUNCTION iam_portal_subject_links_read_v1(
    p_scope_schema name,p_tenant_id uuid,p_principal_id uuid,
    p_token_version bigint,p_authorization_version bigint,
    p_authorization_clock bigint,p_form_id uuid,p_form_version_id uuid,
    p_record_ids uuid[]
)
RETURNS TABLE (
    subject_record_id uuid,portal_user_id uuid,status text,
    account_type text,result_revision bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_requested integer;
BEGIN
    v_requested := pg_catalog.cardinality(p_record_ids);
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS NULL OR p_principal_id IS NULL OR
       p_token_version<1 OR p_authorization_version<1 OR p_authorization_clock<1 OR
       p_form_id IS NULL OR p_form_version_id IS NULL OR
       v_requested IS NULL OR v_requested<1 OR v_requested>500 OR
       EXISTS(SELECT 1 FROM pg_catalog.unnest(p_record_ids) requested(id)
         WHERE requested.id IS NULL) OR
       (SELECT pg_catalog.count(DISTINCT requested.id)
          FROM pg_catalog.unnest(p_record_ids) requested(id))<>v_requested OR
       NOT EXISTS(SELECT 1 FROM tenancy_schema_identity identity_row
          WHERE identity_row.singleton AND identity_row.tenant_id=p_tenant_id
            AND identity_row.schema_name=p_scope_schema::text) OR
       NOT EXISTS(SELECT 1 FROM iam_users principal
          WHERE principal.id=p_principal_id AND principal.status='active'
            AND principal.token_version=p_token_version
            AND principal.authorization_version=p_authorization_version) OR
       NOT EXISTS(SELECT 1 FROM iam_authorization_clock clock
          WHERE clock.singleton AND clock.clock_value=p_authorization_clock) THEN
        RAISE EXCEPTION USING ERRCODE='42501';
    END IF;
    RETURN QUERY
    SELECT binding.subject_record_id,subject.id,
      CASE
        WHEN subject.status<>'active' THEN 'INACTIVE'
        WHEN binding.state='active' THEN 'ACTIVE'
        WHEN binding.state='pending' THEN 'PENDING'
        WHEN setup.approval_state IN('pending','revision_required') THEN 'PENDING'
        ELSE 'INACTIVE'
      END,
      CASE WHEN console.id IS NULL THEN 'portal' ELSE 'operator_portal' END,
      subject.revision
    FROM iam_portal_subject_bindings binding
    JOIN iam_users subject ON subject.id=binding.user_id
      AND subject.status<>'tombstoned'
    JOIN iam_portal_access_setups setup ON setup.binding_id=binding.id
    LEFT JOIN iam_user_access_modes console ON console.user_id=subject.id
      AND console.mode='operator_console' AND console.state='active'
    WHERE binding.profile_form_id=p_form_id
      AND binding.profile_form_version_id=p_form_version_id
      AND binding.subject_record_id=ANY(p_record_ids)
      AND binding.subject_shape='form_record' AND binding.state<>'retired'
    ORDER BY binding.subject_record_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_user_read_v1(name,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_portal_access_verify_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_portal_subject_links_read_v1(
    name,uuid,uuid,bigint,bigint,bigint,uuid,uuid,uuid[]
) FROM PUBLIC;
