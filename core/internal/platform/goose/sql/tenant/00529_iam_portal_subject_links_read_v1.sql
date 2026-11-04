-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_portal_subject_links_read_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_principal_id uuid,
    in_token_version bigint,
    in_authorization_version bigint,
    in_authorization_clock bigint,
    in_form_id uuid,
    in_version_id uuid,
    in_record_i_ds uuid[]
)
RETURNS TABLE (
    subject_record_id uuid,
    user_id uuid,
    status text,
    account_type text,
    revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_subject_links_read_v1(name, uuid, uuid, bigint, bigint, bigint, uuid, uuid, uuid[]) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_portal_subject_links_read_v1(name, uuid, uuid, bigint, bigint, bigint, uuid, uuid, uuid[]) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_portal_subject_links_read_v1(name, uuid, uuid, bigint, bigint, bigint, uuid, uuid, uuid[]);
