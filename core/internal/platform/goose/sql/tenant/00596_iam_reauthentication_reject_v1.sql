-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_reauthentication_reject_v1(
    in_schema_name name,
    in_challenge_id uuid,
    in_expected_revision bigint,
    in_expected_attempt_count smallint,
    in_authenticator_id uuid,
    in_expected_authenticator_revision bigint,
    in_quarantine_authenticator boolean,
    in_expire boolean
)
RETURNS TABLE (
    state text,
    revision bigint,
    attempt_count smallint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        0::bigint,
        0::smallint
    ;
$$;

REVOKE ALL ON FUNCTION iam_reauthentication_reject_v1(name, uuid, bigint, smallint, uuid, bigint, boolean, boolean) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_reauthentication_reject_v1(name, uuid, bigint, smallint, uuid, bigint, boolean, boolean) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_reauthentication_reject_v1(name, uuid, bigint, smallint, uuid, bigint, boolean, boolean);
