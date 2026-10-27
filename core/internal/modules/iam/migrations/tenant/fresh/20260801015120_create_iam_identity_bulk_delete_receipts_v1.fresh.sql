-- +goose Up
CREATE TABLE iam_identity_bulk_delete_receipts (
    operation_id uuid PRIMARY KEY,
    actor_id uuid NOT NULL REFERENCES iam_users(id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    resource_kind text COLLATE "C" NOT NULL
        CHECK (resource_kind IN ('user','role','organization')),
    request_sha256 bytea NOT NULL
        CHECK (pg_catalog.octet_length(request_sha256) = 32),
    deleted_count integer NOT NULL CHECK (deleted_count BETWEEN 1 AND 100),
    created_at timestamptz NOT NULL
);

CREATE TRIGGER iam_identity_bulk_delete_receipts_mutation_guard
BEFORE UPDATE OR DELETE OR TRUNCATE ON iam_identity_bulk_delete_receipts
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_identity_change_mutation_v1();
ALTER TABLE iam_identity_bulk_delete_receipts
    ENABLE ALWAYS TRIGGER iam_identity_bulk_delete_receipts_mutation_guard;

-- +goose StatementBegin
CREATE FUNCTION iam_identity_bulk_delete_reserve_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_resource_kind text, p_request_sha256 bytea, p_deleted_count integer
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_existing iam_identity_bulk_delete_receipts%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_operation_id IS NULL OR p_actor_id IS NULL
       OR p_resource_kind NOT IN ('user','role','organization')
       OR pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32
       OR p_deleted_count NOT BETWEEN 1 AND 100
       OR NOT EXISTS (
           SELECT 1 FROM iam_users AS actor
           WHERE actor.id = p_actor_id AND actor.status = 'active'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM bulk delete receipt scope is invalid';
    END IF;
    INSERT INTO iam_identity_bulk_delete_receipts (
        operation_id, actor_id, resource_kind, request_sha256,
        deleted_count, created_at
    ) VALUES (
        p_operation_id, p_actor_id, p_resource_kind, p_request_sha256,
        p_deleted_count, pg_catalog.transaction_timestamp()
    ) ON CONFLICT (operation_id) DO NOTHING;
    IF FOUND THEN RETURN false; END IF;
    SELECT * INTO STRICT v_existing
    FROM iam_identity_bulk_delete_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id
    FOR SHARE;
    IF v_existing.actor_id IS DISTINCT FROM p_actor_id
       OR v_existing.resource_kind IS DISTINCT FROM p_resource_kind
       OR v_existing.request_sha256 IS DISTINCT FROM p_request_sha256
       OR v_existing.deleted_count IS DISTINCT FROM p_deleted_count THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'IAM bulk delete replay conflicts';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_bulk_delete_reserve_v1(
    name,uuid,uuid,text,bytea,integer
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_identity_bulk_delete_reserve_v1(
    name,uuid,uuid,text,bytea,integer
) TO harpy_runtime;
