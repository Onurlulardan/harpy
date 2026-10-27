-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_portal_user_create_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_requested_user_id uuid,
    p_profile_schema_id uuid, p_profile_document bytea, p_profile_sha256 bytea,
    p_portal_identifier_id uuid, p_portal_display text,
    p_portal_normalized text, p_portal_sha256 bytea,
    p_email_identifier_id uuid, p_email_display text,
    p_email_normalized text, p_email_sha256 bytea,
    p_access_mode_id uuid, p_binding_id uuid, p_subject_type text,
    p_form_id uuid, p_form_version_id uuid, p_record_id uuid,
    p_identifier_field_id uuid, p_contract_sha256 bytea,
    p_initial_status text, p_role_names text[], p_assignment_ids uuid[],
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid, result_revision bigint, authorization_clock bigint,
    replayed boolean, email_identifier_id uuid, email_revision bigint,
    portal_identifier_id uuid, portal_revision bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_user_id uuid; v_values jsonb; v_revision bigint; v_clock bigint;
    v_email_id uuid; v_email_revision bigint; v_portal_id uuid;
    v_portal_revision bigint; v_role_id uuid; v_assignment_id uuid;
    v_console_access_id uuid; v_console_assignment_id uuid; v_console_role_id uuid;
    v_existing_binding iam_portal_subject_bindings%ROWTYPE;
    v_receipt iam_authorization_operation_receipts%ROWTYPE;
    v_index integer; v_state text; v_source text;
BEGIN
    IF p_tenant_id IS NULL OR p_requested_user_id IS NULL OR
       p_portal_identifier_id IS NULL OR p_email_identifier_id IS NULL OR
       p_access_mode_id IS NULL OR p_binding_id IS NULL OR
       p_form_id IS NULL OR p_form_version_id IS NULL OR p_record_id IS NULL OR
       p_identifier_field_id IS NULL OR p_initial_status NOT IN ('PENDING','INACTIVE') OR
       pg_catalog.octet_length(p_portal_sha256) <> 32 OR
       pg_catalog.sha256(pg_catalog.convert_to(p_portal_normalized,'UTF8')) <> p_portal_sha256 OR
       pg_catalog.octet_length(p_email_sha256) <> 32 OR
       pg_catalog.sha256(pg_catalog.convert_to(p_email_normalized,'UTF8')) <> p_email_sha256 OR
       pg_catalog.octet_length(p_contract_sha256) <> 32 OR
       pg_catalog.cardinality(p_role_names) NOT BETWEEN 1 AND 32 OR
       pg_catalog.cardinality(p_role_names) <> pg_catalog.cardinality(p_assignment_ids) OR
       (SELECT pg_catalog.count(DISTINCT role_name) FROM pg_catalog.unnest(p_role_names) role_name)
          <> pg_catalog.cardinality(p_role_names) OR
       NOT EXISTS (SELECT 1 FROM platform.tenants tenant
          WHERE tenant.id = p_tenant_id AND tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM portal user create input is invalid';
    END IF;
    IF NOT iam_identity_actor_assert_v1(
        p_scope_schema,p_actor_id,p_permission_id,
        p_expected_token_version,p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    SELECT receipt.* INTO v_receipt
    FROM iam_authorization_operation_receipts receipt
    WHERE receipt.operation_id = p_operation_id;
    IF FOUND THEN
        IF v_receipt.command_key <> 'iam.portal_user.create' OR
           v_receipt.request_sha256 <> p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505';
        END IF;
        SELECT id,revision INTO STRICT v_email_id,v_email_revision
        FROM iam_user_identifiers WHERE user_id=v_receipt.result_id AND kind='email_v1'
          AND state <> 'retired';
        SELECT id,revision INTO STRICT v_portal_id,v_portal_revision
        FROM iam_user_identifiers WHERE user_id=v_receipt.result_id AND kind='portal_v1'
          AND state <> 'retired';
        RETURN QUERY SELECT v_receipt.result_id,v_receipt.result_revision,
          v_receipt.clock_value,true,v_email_id,v_email_revision,v_portal_id,v_portal_revision;
        RETURN;
    END IF;
    SELECT binding.* INTO v_existing_binding
    FROM iam_portal_subject_bindings binding
    WHERE binding.profile_form_id=p_form_id AND binding.subject_record_id=p_record_id
      AND binding.state <> 'retired' FOR UPDATE;
    IF FOUND THEN
        IF v_existing_binding.subject_type <> p_subject_type OR
           v_existing_binding.profile_form_version_id <> p_form_version_id OR
           v_existing_binding.identifier_field_id <> p_identifier_field_id OR
           v_existing_binding.contract_sha256 <> p_contract_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505';
        END IF;
        v_user_id := v_existing_binding.user_id;
    ELSE
        SELECT identifier.user_id INTO v_user_id
        FROM iam_user_identifiers identifier
        WHERE identifier.kind='email_v1' AND identifier.normalized_value=p_email_normalized
          AND identifier.state <> 'retired' FOR UPDATE;
    END IF;
    IF v_user_id IS NULL THEN
        v_values := iam_identity_profile_assert_v2(
            'user',p_profile_schema_id,p_profile_document,p_profile_sha256
        );
        v_user_id := p_requested_user_id;
        INSERT INTO iam_users(id,identity_kind,status,profile_resource_kind,
          profile_schema_version_id,profile_values,profile_sha256,token_version,
          authorization_version,revision,created_at,updated_at)
        VALUES(v_user_id,'human','active','user',p_profile_schema_id,v_values,
          p_profile_sha256,1,1,1,v_now,v_now);
        INSERT INTO iam_user_identifiers(id,user_id,kind,display_value,normalized_value,
          normalized_sha256,state,revision,verified_at,activated_at,created_at,updated_at)
        VALUES(p_email_identifier_id,v_user_id,'email_v1',p_email_display,p_email_normalized,
          p_email_sha256,'active',1,v_now,v_now,v_now,v_now);
        v_email_id:=p_email_identifier_id; v_email_revision:=1;
    ELSE
        PERFORM 1 FROM iam_users subject
        WHERE subject.id=v_user_id AND subject.identity_kind='human'
          AND subject.status='active' FOR UPDATE;
        IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='23505'; END IF;
        SELECT id,revision INTO STRICT v_email_id,v_email_revision
        FROM iam_user_identifiers WHERE user_id=v_user_id AND kind='email_v1'
          AND normalized_value=p_email_normalized AND state='active';
    END IF;
    IF v_existing_binding.id IS NULL THEN
        INSERT INTO iam_user_identifiers(id,user_id,kind,display_value,normalized_value,
          normalized_sha256,state,revision,created_at,updated_at)
        VALUES(p_portal_identifier_id,v_user_id,'portal_v1',p_portal_display,
          p_portal_normalized,p_portal_sha256,'pending',1,v_now,v_now);
        v_portal_id:=p_portal_identifier_id; v_portal_revision:=1;
        v_state:=CASE p_initial_status WHEN 'PENDING' THEN 'pending' ELSE 'suspended' END;
        INSERT INTO iam_user_access_modes(id,user_id,mode,state,revision,created_at,
          updated_at,suspended_at)
        VALUES(p_access_mode_id,v_user_id,'portal',v_state,1,v_now,v_now,
          CASE WHEN v_state='suspended' THEN v_now END);
        INSERT INTO iam_portal_subject_bindings(id,user_id,access_mode_id,access_mode,
          subject_shape,subject_type,profile_owner,profile_form_id,profile_form_version_id,
          subject_record_id,identifier_field_id,contract_sha256,state,revision,created_at,
          updated_at,suspended_at)
        VALUES(p_binding_id,v_user_id,p_access_mode_id,'portal','form_record',p_subject_type,
          'forms',p_form_id,p_form_version_id,p_record_id,p_identifier_field_id,
          p_contract_sha256,v_state,1,v_now,v_now,CASE WHEN v_state='suspended' THEN v_now END);
        INSERT INTO iam_portal_access_setups(binding_id,user_id,access_mode_id,access_mode,
          setup_state,approval_state,provenance,approval_revision,source_evidence_sha256,
          revision,created_at,updated_at)
        VALUES(p_binding_id,v_user_id,p_access_mode_id,'portal','required','not_required',
          'native_v2',0,p_contract_sha256,1,v_now,v_now);
    ELSE
        SELECT id,revision INTO STRICT v_portal_id,v_portal_revision
        FROM iam_user_identifiers WHERE user_id=v_user_id AND kind='portal_v1'
          AND normalized_value=p_portal_normalized AND state <> 'retired';
    END IF;
    FOR v_index IN 1..pg_catalog.cardinality(p_role_names) LOOP
        SELECT role.id INTO v_role_id FROM iam_roles role
        WHERE role.scope_kind='tenant' AND role.scope_id=p_tenant_id
          AND role.status='active' AND role.name_key=pg_catalog.convert_to(p_role_names[v_index],'UTF8');
        IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='23503'; END IF;
        SELECT assignment.id INTO v_assignment_id FROM iam_user_role_assignments assignment
        WHERE assignment.user_id=v_user_id AND assignment.role_id=v_role_id FOR UPDATE;
        IF FOUND THEN
            UPDATE iam_user_role_assignments SET state='active',revoked_at=NULL,
              revision=revision+1,updated_at=v_now WHERE id=v_assignment_id AND state<>'active';
        ELSE
            v_assignment_id:=p_assignment_ids[v_index];
            INSERT INTO iam_user_role_assignments(id,user_id,role_id,role_scope_kind,
              role_scope_id,state,revision,created_at,updated_at)
            VALUES(v_assignment_id,v_user_id,v_role_id,'tenant',p_tenant_id,'active',1,v_now,v_now);
        END IF;
        INSERT INTO iam_portal_role_assignments(binding_id,user_id,assignment_id,role_id,created_at)
        VALUES(COALESCE(v_existing_binding.id,p_binding_id),v_user_id,v_assignment_id,v_role_id,v_now)
        ON CONFLICT (binding_id,role_id) DO NOTHING;
    END LOOP;
    SELECT access.id,assignment.id,assignment.role_id
    INTO v_console_access_id,v_console_assignment_id,v_console_role_id
    FROM iam_user_access_modes access
    JOIN iam_user_role_assignments assignment ON assignment.user_id=access.user_id
    JOIN iam_roles role ON role.id=assignment.role_id AND role.is_default
    WHERE access.user_id=v_user_id AND access.mode='operator_console' AND access.state='active'
      AND assignment.state='active' LIMIT 1;
    IF FOUND THEN
        INSERT INTO iam_portal_console_grants(user_id,access_mode_id,access_mode,
          assignment_id,role_id,source,state,revision,created_at,updated_at)
        VALUES(v_user_id,v_console_access_id,'operator_console',v_console_assignment_id,
          v_console_role_id,
          'preexisting_operator','active',1,v_now,v_now) ON CONFLICT(user_id) DO NOTHING;
    END IF;
    UPDATE iam_users SET token_version=token_version+1,
      authorization_version=authorization_version+1,revision=revision+1,updated_at=v_now
    WHERE id=v_user_id AND id<>p_requested_user_id RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN SELECT revision INTO v_revision FROM iam_users WHERE id=v_user_id; END IF;
    SELECT id,revision INTO STRICT v_portal_id,v_portal_revision
    FROM iam_user_identifiers WHERE user_id=v_user_id AND kind='portal_v1' AND state<>'retired';
    v_clock:=iam_identity_record_v1(p_operation_id,'iam.portal_user.create',
      p_request_sha256,v_user_id,v_revision,p_request_sha256);
    RETURN QUERY SELECT v_user_id,v_revision,v_clock,false,
      v_email_id,v_email_revision,v_portal_id,v_portal_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_user_create_v1(
    name,uuid,uuid,bytea,uuid,uuid,bytea,bytea,uuid,text,text,bytea,
    uuid,text,text,bytea,uuid,uuid,text,uuid,uuid,uuid,uuid,bytea,
    text,text[],uuid[],uuid,uuid,bigint,bigint
) FROM PUBLIC;
