-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_post_restore_reactivate_owner_v1(
    in_operation_id uuid,
    in_environment_id uuid,
    in_restore_operation_id uuid,
    in_tenant_id uuid,
    in_owner_user_id uuid,
    in_anchor_id uuid,
    in_assertion_sha256 bytea,
    in_password_sha256 bytea,
    in_receipt_sha256 bytea,
    in_sign_count bigint,
    in_salt bytea,
    in_tag bytea,
    in_memory_ki_b integer,
    in_iterations smallint,
    in_parallelism smallint,
    in_id uuid,
    in_revision bigint,
    in_directory_key_id text,
    in_directory_hmac bytea,
    in_occurred_at timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION iam_post_restore_reactivate_owner_v1(uuid, uuid, uuid, uuid, uuid, uuid, bytea, bytea, bytea, bigint, bytea, bytea, integer, smallint, smallint, uuid, bigint, text, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_post_restore_reactivate_owner_v1(uuid, uuid, uuid, uuid, uuid, uuid, bytea, bytea, bytea, bigint, bytea, bytea, integer, smallint, smallint, uuid, bigint, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_post_restore_reactivate_owner_v1(uuid, uuid, uuid, uuid, uuid, uuid, bytea, bytea, bytea, bigint, bytea, bytea, integer, smallint, smallint, uuid, bigint, text, bytea, timestamptz);
