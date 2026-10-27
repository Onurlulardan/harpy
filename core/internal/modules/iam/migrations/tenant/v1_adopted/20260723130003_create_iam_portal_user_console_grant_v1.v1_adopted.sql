-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_portal_user_console_grant_v1(
    p_scope_schema name,p_tenant_id uuid,p_operation_id uuid,
    p_request_sha256 bytea,p_user_id uuid,p_expected_revision bigint,
    p_access_mode_id uuid,p_assignment_id uuid,p_actor_id uuid,
    p_permission_id uuid,p_expected_token_version bigint,
    p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid,result_revision bigint,authorization_clock bigint,
    replayed boolean,email_identifier_id uuid,email_revision bigint,
    portal_identifier_id uuid,portal_revision bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz:=pg_catalog.transaction_timestamp();
    v_role_id uuid; v_access_id uuid; v_assignment_id uuid; v_revision bigint;
    v_clock bigint; v_email_id uuid; v_email_revision bigint;
    v_portal_id uuid; v_portal_revision bigint; v_grant iam_portal_console_grants%ROWTYPE;
BEGIN
    IF p_user_id IS NULL OR p_expected_revision<1 OR p_access_mode_id IS NULL OR
       p_assignment_id IS NULL OR NOT EXISTS(SELECT 1 FROM platform.tenants tenant
         WHERE tenant.id=p_tenant_id AND tenant.schema_name=p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE='22023';
    END IF;
    RETURN QUERY SELECT replay.result_id,replay.result_revision,replay.authorization_clock,
      replay.replayed,email.id,email.revision,portal.id,portal.revision
    FROM iam_identity_replay_v1(p_operation_id,'iam.portal_user.console.grant',
      p_request_sha256,p_user_id) replay
    JOIN iam_user_identifiers email ON email.user_id=replay.result_id
      AND email.kind='email_v1' AND email.state<>'retired'
    JOIN iam_user_identifiers portal ON portal.user_id=replay.result_id
      AND portal.kind='portal_v1' AND portal.state<>'retired';
    IF FOUND THEN RETURN; END IF;
    IF NOT iam_identity_actor_assert_v1(
      p_scope_schema,p_actor_id,p_permission_id,
      p_expected_token_version,p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE='42501'; END IF;
    PERFORM 1 FROM iam_users subject
    WHERE subject.id=p_user_id AND subject.status='active'
      AND subject.revision=p_expected_revision FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='40001'; END IF;
    PERFORM 1 FROM iam_portal_subject_bindings binding
    JOIN iam_portal_access_setups setup ON setup.binding_id=binding.id
    JOIN iam_password_credentials credential ON credential.user_id=binding.user_id
    JOIN iam_user_identifiers email ON email.user_id=binding.user_id
      AND email.kind='email_v1' AND email.state='active'
    WHERE binding.user_id=p_user_id AND binding.state<>'retired'
      AND setup.setup_state='completed' AND credential.state='active' FOR SHARE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='23514',
      MESSAGE='IAM portal console prerequisites are incomplete'; END IF;
    SELECT role.id INTO v_role_id FROM iam_roles role
    WHERE role.scope_kind='tenant' AND role.scope_id=p_tenant_id
      AND role.status='active' AND role.is_default FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='23503'; END IF;
    SELECT grant_row.* INTO v_grant FROM iam_portal_console_grants grant_row
      WHERE grant_row.user_id=p_user_id FOR UPDATE;
    IF FOUND THEN
        v_access_id:=v_grant.access_mode_id;
        v_assignment_id:=v_grant.assignment_id;
        UPDATE iam_user_access_modes SET state='active',
          activated_at=COALESCE(activated_at,v_now),suspended_at=NULL,
          revision=revision+1,updated_at=v_now
        WHERE id=v_access_id AND state<>'active';
        UPDATE iam_user_role_assignments SET state='active',revoked_at=NULL,
          revision=revision+1,updated_at=v_now
        WHERE id=v_assignment_id AND state<>'active';
        UPDATE iam_portal_console_grants SET state='active',revoked_at=NULL,
          revision=revision+1,updated_at=v_now WHERE user_id=p_user_id AND state<>'active';
    ELSE
        SELECT access.id INTO v_access_id FROM iam_user_access_modes access
        WHERE access.user_id=p_user_id AND access.mode='operator_console' FOR UPDATE;
        IF FOUND THEN
            UPDATE iam_user_access_modes SET state='active',
              activated_at=COALESCE(activated_at,v_now),suspended_at=NULL,
              revision=revision+1,updated_at=v_now WHERE id=v_access_id AND state<>'active';
        ELSE
            v_access_id:=p_access_mode_id;
            INSERT INTO iam_user_access_modes(id,user_id,mode,state,revision,
              created_at,updated_at,activated_at)
            VALUES(v_access_id,p_user_id,'operator_console','active',1,v_now,v_now,v_now);
        END IF;
        SELECT assignment.id INTO v_assignment_id FROM iam_user_role_assignments assignment
        WHERE assignment.user_id=p_user_id AND assignment.role_id=v_role_id FOR UPDATE;
        IF FOUND THEN
            UPDATE iam_user_role_assignments SET state='active',revoked_at=NULL,
              revision=revision+1,updated_at=v_now
            WHERE id=v_assignment_id AND state<>'active';
        ELSE
            v_assignment_id:=p_assignment_id;
            INSERT INTO iam_user_role_assignments(id,user_id,role_id,role_scope_kind,
              role_scope_id,state,revision,created_at,updated_at)
            VALUES(v_assignment_id,p_user_id,v_role_id,'tenant',p_tenant_id,
              'active',1,v_now,v_now);
        END IF;
        INSERT INTO iam_portal_console_grants(user_id,access_mode_id,access_mode,
          assignment_id,role_id,source,state,revision,created_at,updated_at)
        VALUES(p_user_id,v_access_id,'operator_console',v_assignment_id,v_role_id,
          'portal_grant','active',1,v_now,v_now);
    END IF;
    UPDATE iam_users SET token_version=token_version+1,
      authorization_version=authorization_version+1,revision=revision+1,updated_at=v_now
    WHERE id=p_user_id RETURNING revision INTO v_revision;
    SELECT id,revision INTO STRICT v_email_id,v_email_revision FROM iam_user_identifiers
      WHERE user_id=p_user_id AND kind='email_v1' AND state<>'retired';
    SELECT id,revision INTO STRICT v_portal_id,v_portal_revision FROM iam_user_identifiers
      WHERE user_id=p_user_id AND kind='portal_v1' AND state<>'retired';
    v_clock:=iam_identity_record_v1(p_operation_id,'iam.portal_user.console.grant',
      p_request_sha256,p_user_id,v_revision,p_request_sha256);
    RETURN QUERY SELECT p_user_id,v_revision,v_clock,false,
      v_email_id,v_email_revision,v_portal_id,v_portal_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_user_console_grant_v1(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;

