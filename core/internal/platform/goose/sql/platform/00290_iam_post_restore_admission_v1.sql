-- +goose Up
CREATE FUNCTION platform.iam_post_restore_admission_v1(
    in_arg boolean
)
RETURNS TABLE (
    state text,
    environment_id uuid,
    restore_operation_id uuid
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid
    ;
$$;

REVOKE ALL ON FUNCTION platform.iam_post_restore_admission_v1(boolean) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_post_restore_admission_v1(boolean) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_post_restore_admission_v1(boolean);
