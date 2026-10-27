-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_import_receipts_v1(
    p_scope_schema name,p_import_id uuid,p_actor_id uuid,p_resource_kind text,
    p_token_version bigint,p_authorization_version bigint,
    p_after_row bigint,p_limit integer
)
RETURNS TABLE(
    row_index bigint,outcome text,owner_kind text,owner_id uuid,error_code text,
    source_sha256 bytea,result_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_import_exists boolean;
BEGIN
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'import',
        p_token_version,p_authorization_version,false
    ) THEN NULL; END IF;
    IF p_after_row NOT BETWEEN 0 AND 1000000 OR
       p_limit NOT BETWEEN 1 AND 101 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM import receipt page is invalid';
    END IF;
    SELECT EXISTS(
        SELECT 1 FROM iam_system_record_imports AS request
        WHERE request.import_id=p_import_id AND request.actor_id=p_actor_id
          AND request.resource_kind=p_resource_kind
    ) INTO v_import_exists;
    IF NOT v_import_exists THEN
        RAISE EXCEPTION USING ERRCODE='P0002',
            MESSAGE='IAM import is unavailable';
    END IF;
    RETURN QUERY
    SELECT receipt.row_index,receipt.outcome,receipt.owner_kind,
        receipt.owner_id,receipt.error_code,receipt.source_sha256,
        receipt.result_sha256
    FROM iam_system_import_row_receipts AS receipt
    WHERE receipt.import_id=p_import_id AND receipt.row_index>p_after_row
    ORDER BY receipt.row_index
    LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_import_receipts_v1(
    name,uuid,uuid,text,bigint,bigint,bigint,integer
) FROM PUBLIC;
