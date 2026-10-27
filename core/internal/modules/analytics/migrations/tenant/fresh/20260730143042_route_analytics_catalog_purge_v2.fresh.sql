-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_component_apply_v2(
    p_receipt_id uuid,p_operation_id uuid,p_component_id uuid,
    p_owner text,p_generation bigint,p_contract text,p_content_sha256 text,
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
        RETURN QUERY
        SELECT result.receipt_id,result.operation_id,result.component_id,
            result.owner_module,result.owner_generation,result.contract,
            result.content_sha256,result.action,result.ready,
            result.completed_at
        FROM analytics_catalog_component_purge_v2(
            p_receipt_id,p_operation_id,p_component_id,p_owner,p_generation,
            p_contract,p_content_sha256,p_action,p_reference_epoch,
            p_idempotency_key,p_completed_at
        ) AS result;
        RETURN;
    END IF;
    RETURN QUERY
    SELECT result.receipt_id,result.operation_id,result.component_id,
        result.owner_module,result.owner_generation,result.contract,
        result.content_sha256,result.action,result.ready,result.completed_at
    FROM analytics_catalog_component_apply_core_v2(
        p_receipt_id,p_operation_id,p_component_id,p_owner,p_generation,
        p_contract,p_content_sha256,p_action,p_reference_epoch,
        p_idempotency_key,p_completed_at
    ) AS result;
END;
$function$;
REVOKE ALL ON FUNCTION analytics_catalog_component_apply_v2(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) FROM PUBLIC;
-- +goose StatementEnd
