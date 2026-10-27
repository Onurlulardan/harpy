-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_recovery_codes_list_v1(
    p_scope_schema name, p_user_id uuid, p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    authenticator_id uuid, authenticator_revision bigint, code_id uuid,
    code_position smallint, state text, created_at timestamptz,
    consumed_at timestamptz, revoked_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_user_id IS NULL OR p_user_id <> p_actor_id OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY
    SELECT authenticator.id, authenticator.revision, code.id,
        code.position, code.state, code.created_at,
        code.consumed_at, code.revoked_at
    FROM iam_authenticators AS authenticator
    JOIN iam_recovery_codes AS code
      ON code.authenticator_id = authenticator.id
     AND code.user_id = authenticator.user_id
    WHERE authenticator.user_id = p_user_id
      AND authenticator.kind = 'recovery_set_v1'
      AND authenticator.state <> 'retired'
    ORDER BY code.position ASC;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_recovery_codes_list_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
