-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identifier_tombstone_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_identifier_id uuid, p_expected_revision bigint,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid, subject_user_id uuid, result_revision bigint,
    token_version bigint, authorization_version bigint, replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_user_id uuid;
    v_state text;
    v_revision bigint;
    v_token_version bigint;
    v_authorization_version bigint;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF p_identifier_id IS NULL OR p_expected_revision < 1 OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND
                tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.identifier.tombstone',
        p_request_sha256, p_identifier_id
    );
    IF FOUND THEN RETURN; END IF;
    IF NOT iam_identity_actor_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    SELECT identifier.user_id, identifier.state, identifier.revision
    INTO v_user_id, v_state, v_revision
    FROM iam_user_identifiers AS identifier
    WHERE identifier.id = p_identifier_id FOR UPDATE;
    IF NOT FOUND OR v_state <> 'active' OR v_revision <> p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM identifier revision changed';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM iam_user_identifiers AS alternative
        WHERE alternative.user_id = v_user_id
          AND alternative.state = 'active' AND alternative.id <> p_identifier_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM user requires another active identifier';
    END IF;
    UPDATE iam_user_identifiers SET state = 'tombstoned',
        tombstoned_at = v_now, cooldown_until = v_now + INTERVAL '30 days',
        revision = revision + 1, updated_at = v_now
    WHERE id = p_identifier_id RETURNING revision INTO v_revision;
    UPDATE iam_one_time_tokens SET state = 'revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE identifier_id = p_identifier_id AND state = 'pending';
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.identifier.tombstone', p_request_sha256,
        p_identifier_id, v_user_id, v_revision,
        'actor_tenant', p_actor_id, 'identifier', true, true
    ) AS result;
    RETURN QUERY SELECT p_identifier_id, v_user_id, v_revision,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identifier_tombstone_v1(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) FROM PUBLIC;
