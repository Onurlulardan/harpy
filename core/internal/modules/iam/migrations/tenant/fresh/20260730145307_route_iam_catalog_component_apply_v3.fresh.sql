-- +goose Up
-- +goose StatementBegin
DROP FUNCTION iam_catalog_component_apply_v1(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
);

CREATE FUNCTION iam_catalog_component_apply_v1(
    p_receipt_id uuid,p_operation_id uuid,p_component_id uuid,p_owner text,
    p_generation bigint,p_contract text,p_content_sha256 text,
    p_action text,p_reference_epoch bigint,p_idempotency_key text,
    p_completed_at timestamptz
)
RETURNS TABLE(
    receipt_id uuid,operation_id uuid,component_id uuid,owner_module text,
    owner_generation bigint,contract text,content_sha256 text,action text,
    ready boolean,completed_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_action='purge' THEN
        RETURN QUERY SELECT result.*
        FROM iam_catalog_component_purge_v3(
            p_receipt_id,p_operation_id,p_component_id,p_owner,p_generation,
            p_contract,p_content_sha256,p_action,p_reference_epoch,
            p_idempotency_key,p_completed_at
        ) AS result;
    ELSIF p_action IN ('disable','enable','retire','restore') THEN
        RETURN QUERY SELECT result.*
        FROM iam_catalog_component_transition_v3(
            p_receipt_id,p_operation_id,p_component_id,p_owner,p_generation,
            p_contract,p_content_sha256,p_action,p_reference_epoch,
            p_idempotency_key,p_completed_at
        ) AS result;
    ELSIF p_action IN ('stage','cleanup','activate') AND
          p_reference_epoch=0 THEN
        RETURN QUERY SELECT result.*
        FROM iam_catalog_component_apply_v2(
            p_receipt_id,p_operation_id,p_component_id,p_owner,p_generation,
            p_contract,p_content_sha256,p_action,p_reference_epoch,
            p_idempotency_key,p_completed_at
        ) AS result;
    ELSE
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM Catalog component action is invalid';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_catalog_component_apply_v1(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_catalog_component_apply_v1(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) TO harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_catalog_component_apply_v2(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) FROM harpy_runtime;
