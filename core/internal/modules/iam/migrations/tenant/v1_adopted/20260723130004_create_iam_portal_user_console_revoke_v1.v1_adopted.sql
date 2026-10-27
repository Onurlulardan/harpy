-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_portal_user_console_revoke_v1(
    p_scope_schema name,p_tenant_id uuid,p_operation_id uuid,
    p_request_sha256 bytea,p_user_id uuid,p_expected_revision bigint,
    p_actor_id uuid,p_permission_id uuid,p_expected_token_version bigint,
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
    v_access_id uuid; v_assignment_id uuid; v_revision bigint; v_clock bigint;
    v_email_id uuid; v_email_revision bigint; v_portal_id uuid; v_portal_revision bigint;
BEGIN
    IF p_user_id IS NULL OR p_expected_revision<1 OR
       NOT EXISTS(SELECT 1 FROM platform.tenants tenant
         WHERE tenant.id=p_tenant_id AND tenant.schema_name=p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE='22023';
    END IF;
    RETURN QUERY SELECT replay.result_id,replay.result_revision,replay.authorization_clock,
      replay.replayed,email.id,email.revision,portal.id,portal.revision
    FROM iam_identity_replay_v1(p_operation_id,'iam.portal_user.console.revoke',
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
    SELECT grant_row.access_mode_id,grant_row.assignment_id
    INTO v_access_id,v_assignment_id
    FROM iam_portal_console_grants grant_row
    WHERE grant_row.user_id=p_user_id AND grant_row.state='active' FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='P0002'; END IF;
    UPDATE iam_refresh_families family SET state='revoked',
      revoke_reason='portal_console_revoked',revoked_at=v_now,
      revision=revision+1,updated_at=v_now
    FROM iam_sessions session WHERE family.session_id=session.id
      AND session.user_id=p_user_id AND session.access_mode_id=v_access_id
      AND family.state='active';
    UPDATE iam_sessions SET state='revoked',revoked_at=v_now,
      revision=revision+1,updated_at=v_now
    WHERE user_id=p_user_id AND access_mode_id=v_access_id AND state='active';
    UPDATE iam_user_access_modes SET state='suspended',suspended_at=v_now,
      revision=revision+1,updated_at=v_now WHERE id=v_access_id;
    IF NOT EXISTS(SELECT 1 FROM iam_portal_role_assignments portal_role
       WHERE portal_role.assignment_id=v_assignment_id) THEN
        UPDATE iam_user_role_assignments SET state='revoked',revoked_at=v_now,
          revision=revision+1,updated_at=v_now WHERE id=v_assignment_id;
    END IF;
    UPDATE iam_portal_console_grants SET state='revoked',revoked_at=v_now,
      revision=revision+1,updated_at=v_now WHERE user_id=p_user_id;
    UPDATE iam_users SET token_version=token_version+1,
      authorization_version=authorization_version+1,revision=revision+1,updated_at=v_now
    WHERE id=p_user_id RETURNING revision INTO v_revision;
    SELECT id,revision INTO STRICT v_email_id,v_email_revision FROM iam_user_identifiers
      WHERE user_id=p_user_id AND kind='email_v1' AND state<>'retired';
    SELECT id,revision INTO STRICT v_portal_id,v_portal_revision FROM iam_user_identifiers
      WHERE user_id=p_user_id AND kind='portal_v1' AND state<>'retired';
    v_clock:=iam_identity_record_v1(p_operation_id,'iam.portal_user.console.revoke',
      p_request_sha256,p_user_id,v_revision,p_request_sha256);
    RETURN QUERY SELECT p_user_id,v_revision,v_clock,false,
      v_email_id,v_email_revision,v_portal_id,v_portal_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_user_console_revoke_v1(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) FROM PUBLIC;

