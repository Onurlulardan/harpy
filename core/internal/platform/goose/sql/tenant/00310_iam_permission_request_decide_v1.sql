-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_decide_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_group_id uuid,
    in_expected_revision bigint,
    in_expected_state text,
    in_reviewer_id uuid,
    in_expected_reviewer_token_version bigint,
    in_expected_reviewer_authorization_version bigint,
    in_review_task_id uuid,
    in_decision text,
    in_purpose_sha256 bytea,
    in_auth_time timestamptz,
    in_acr text,
    in_amr_password boolean,
    in_amr_web_authn boolean,
    in_amrtotp boolean
)
RETURNS TABLE (
    result_code text,
    result_revision bigint,
    authorization_clock bigint,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        0::bigint,
        0::bigint,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_decide_v1(name, uuid, bytea, uuid, bigint, text, uuid, bigint, bigint, uuid, text, bytea, timestamptz, text, boolean, boolean, boolean) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_permission_request_decide_v1(name, uuid, bytea, uuid, bigint, text, uuid, bigint, bigint, uuid, text, bytea, timestamptz, text, boolean, boolean, boolean) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_permission_request_decide_v1(name, uuid, bytea, uuid, bigint, text, uuid, bigint, bigint, uuid, text, bytea, timestamptz, text, boolean, boolean, boolean);
