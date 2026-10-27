-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_owner_transfer_commit_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_actor_id uuid, p_expected_token_version bigint,
    p_expected_authorization_version bigint, p_expected_platform_clock bigint,
    p_outbox_event_id uuid, p_event_sha256 bytea
)
RETURNS TABLE (
    operation_id uuid, predecessor_user_id uuid, successor_user_id uuid,
    predecessor_token_version bigint, predecessor_authorization_version bigint,
    successor_token_version bigint, successor_authorization_version bigint,
    owner_policy_revision bigint, platform_clock bigint, head_generation bigint,
    replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid; v_predecessor_id uuid; v_successor_id uuid;
    v_predecessor_token bigint; v_predecessor_auth bigint;
    v_successor_token bigint; v_successor_auth bigint;
    v_predecessor_revision bigint; v_successor_revision bigint;
    v_policy_revision bigint; v_clock bigint; v_generation bigint; v_replayed boolean;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_operation_id IS NULL OR pg_catalog.octet_length(p_request_sha256) <> 32 OR
       p_actor_id IS NULL OR p_expected_token_version < 1 OR
       p_expected_authorization_version < 1 OR p_expected_platform_clock < 1 OR
       p_outbox_event_id IS NULL OR pg_catalog.octet_length(p_event_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Owner transfer commit input is invalid';
    END IF;
    SELECT identity.tenant_id INTO v_tenant_id
    FROM tenancy_schema_identity AS identity WHERE identity.singleton;
    PERFORM 1 FROM iam_users AS principal
    WHERE principal.id = p_actor_id AND principal.status = 'active'
      AND principal.identity_kind = 'human'
      AND principal.token_version IN (p_expected_token_version,p_expected_token_version + 1)
      AND principal.authorization_version IN (
          p_expected_authorization_version,p_expected_authorization_version + 1
      ) FOR UPDATE;
    IF NOT FOUND OR v_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer commit actor is stale';
    END IF;
    SELECT result.predecessor_user_id,result.successor_user_id,
           result.predecessor_token_version,result.predecessor_authorization_version,
           result.successor_token_version,result.successor_authorization_version,
           result.predecessor_expected_revision,result.successor_expected_revision,
           result.owner_policy_revision,result.platform_clock,result.head_generation,result.replayed
    INTO v_predecessor_id,v_successor_id,v_predecessor_token,v_predecessor_auth,
         v_successor_token,v_successor_auth,v_predecessor_revision,v_successor_revision,
         v_policy_revision,v_clock,v_generation,v_replayed
    FROM platform.iam_owner_transfer_commit_v1(
        v_tenant_id,p_operation_id,p_request_sha256,p_actor_id,p_expected_platform_clock
    ) AS result;
    IF v_replayed THEN
        IF NOT EXISTS (SELECT 1 FROM iam_users AS principal
            WHERE principal.id = v_predecessor_id AND principal.token_version = v_predecessor_token
              AND principal.authorization_version = v_predecessor_auth) OR
           NOT EXISTS (SELECT 1 FROM iam_users AS principal
            WHERE principal.id = v_successor_id AND principal.token_version = v_successor_token
              AND principal.authorization_version = v_successor_auth) OR
           NOT EXISTS (SELECT 1 FROM iam_platform_owner_transfer_audit AS audit
            WHERE audit.operation_id = p_operation_id AND audit.request_sha256 = p_request_sha256
              AND audit.platform_clock = v_clock AND audit.head_generation = v_generation) OR
           NOT EXISTS (SELECT 1 FROM iam_platform_owner_transfer_outbox AS event
            WHERE event.operation_id = p_operation_id AND event.event_id = p_outbox_event_id
              AND event.payload_sha256 = p_event_sha256) THEN
            RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer replay is stale';
        END IF;
    ELSE
        UPDATE iam_users SET token_version = v_predecessor_token,
            authorization_version = v_predecessor_auth, revision = revision + 1, updated_at = v_now
        WHERE id = v_predecessor_id AND revision = v_predecessor_revision
          AND token_version = v_predecessor_token - 1
          AND authorization_version = v_predecessor_auth - 1;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner predecessor CAS failed';
        END IF;
        UPDATE iam_users SET token_version = v_successor_token,
            authorization_version = v_successor_auth, revision = revision + 1, updated_at = v_now
        WHERE id = v_successor_id AND revision = v_successor_revision
          AND token_version = v_successor_token - 1
          AND authorization_version = v_successor_auth - 1;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner successor CAS failed';
        END IF;
        UPDATE iam_sessions SET state = 'revoked', revoked_at = v_now,
            updated_at = v_now, revision = revision + 1
        WHERE user_id IN (v_predecessor_id,v_successor_id) AND state = 'active';
        INSERT INTO iam_platform_owner_transfer_audit VALUES (
            p_operation_id,v_predecessor_id,v_successor_id,p_request_sha256,
            v_policy_revision,v_clock,v_generation,v_now
        );
        INSERT INTO iam_platform_owner_transfer_outbox VALUES (
            p_outbox_event_id,p_operation_id,'iam.platform.owner_transferred.v1',
            p_event_sha256,'pending',v_now
        );
    END IF;
    RETURN QUERY SELECT p_operation_id,v_predecessor_id,v_successor_id,
        v_predecessor_token,v_predecessor_auth,v_successor_token,v_successor_auth,
        v_policy_revision,v_clock,v_generation,v_replayed;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_owner_transfer_commit_v1(
    name,uuid,bytea,uuid,bigint,bigint,bigint,uuid,bytea
) FROM PUBLIC;
