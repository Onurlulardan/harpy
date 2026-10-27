-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_project_v1(
    p_scope_schema name,p_receipt_id uuid,
    p_package_id uuid,p_package_slug text,p_package_version text,
    p_release_id uuid,p_component_id uuid,p_owner_generation bigint,
    p_contract text,p_content_sha256 text,
    p_component_kind text,p_module_slug text,p_semantic_identity text,
    p_semantic_sha256 bytea,p_item_count integer,
    p_auxiliary_count integer,p_document_bytes bytea,
    p_idempotency_key text,p_request_sha256 bytea,
    p_occurred_at timestamptz
)
RETURNS TABLE(
    receipt_id uuid,package_id uuid,package_slug text,
    package_version text,release_id uuid,component_id uuid,
    owner_generation bigint,contract text,content_sha256 text,
    component_kind text,module_slug text,semantic_identity text,
    semantic_sha256 bytea,item_count integer,auxiliary_count integer,
    document_bytes bytea,idempotency_key text,request_sha256 bytea,
    lifecycle_state text,ready_at timestamptz,
    created_at timestamptz,updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_document jsonb;
    v_receipt iam_catalog_provisions%ROWTYPE;
BEGIN
    IF p_scope_schema IS DISTINCT FROM (
           SELECT identity_row.schema_name::name
           FROM tenancy_schema_identity AS identity_row
           WHERE identity_row.singleton
       ) OR
       pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       p_package_id IS NULL OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       p_package_version IS NULL OR
       p_package_version !~
        '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$' OR
       p_owner_generation<1 OR
       length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_idempotency_key !~ '^[A-Za-z0-9_.:-]+$' OR
       octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) OR
       date_trunc('microseconds',p_occurred_at)<>p_occurred_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM catalog projection request is invalid';
    END IF;
    v_document := iam_catalog_project_document_v1(
        p_package_slug,p_component_id,p_contract,p_content_sha256,
        p_component_kind,p_module_slug,p_semantic_identity,
        p_semantic_sha256,p_item_count,p_auxiliary_count,p_document_bytes
    );
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'iam:catalog-project:'||p_component_id::text,6201
    ));
    SELECT * INTO v_receipt FROM iam_catalog_provisions AS provision
    WHERE provision.component_id=p_component_id
      AND provision.idempotency_key=p_idempotency_key
    FOR UPDATE;
    IF FOUND THEN
        IF ROW(v_receipt.package_id,v_receipt.package_slug,
               v_receipt.package_version,v_receipt.release_id,
               v_receipt.owner_generation,v_receipt.contract,
               v_receipt.content_sha256,v_receipt.component_kind,
               v_receipt.module_slug,v_receipt.semantic_identity,
               v_receipt.semantic_sha256,v_receipt.item_count,
               v_receipt.auxiliary_count,v_receipt.request_sha256)
           IS DISTINCT FROM
           ROW(p_package_id,p_package_slug,p_package_version,p_release_id,
               p_owner_generation,p_contract,p_content_sha256,
               p_component_kind,p_module_slug,p_semantic_identity,
               p_semantic_sha256,p_item_count,p_auxiliary_count,
               p_request_sha256) THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='IAM catalog projection replay differs';
        END IF;
        RETURN QUERY SELECT * FROM
            iam_catalog_projection_read_v1(v_receipt.receipt_id);
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='IAM catalog replay readback differs';
        END IF;
        RETURN;
    END IF;
    PERFORM iam_catalog_materialize_v1(
        p_package_id,p_package_slug,p_package_version,
        p_release_id,p_component_id,p_owner_generation,
        p_contract,p_content_sha256,p_component_kind,
        p_module_slug,p_semantic_sha256,p_document_bytes,
        v_document,p_occurred_at
    );
    PERFORM iam_catalog_operational_materialize_v1(
        p_scope_schema,p_component_id,v_document,
        p_semantic_sha256,p_occurred_at
    );
    IF NOT iam_catalog_operational_readback_v1(p_component_id) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM catalog operational readback differs';
    END IF;
    INSERT INTO iam_catalog_provisions(
        receipt_id,package_id,package_slug,package_version,release_id,
        component_id,owner_generation,contract,content_sha256,
        component_kind,module_slug,semantic_identity,semantic_sha256,
        item_count,auxiliary_count,idempotency_key,request_sha256,
        ready_at,created_at,updated_at
    ) VALUES (
        p_receipt_id,p_package_id,p_package_slug,p_package_version,
        p_release_id,p_component_id,p_owner_generation,p_contract,
        p_content_sha256,p_component_kind,p_module_slug,
        p_semantic_identity,p_semantic_sha256,p_item_count,
        p_auxiliary_count,p_idempotency_key,p_request_sha256,
        p_occurred_at,p_occurred_at,p_occurred_at
    );
    RETURN QUERY SELECT * FROM iam_catalog_projection_read_v1(p_receipt_id);
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM catalog projection readback differs';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_catalog_project_v1(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,text,bytea,timestamptz
) FROM PUBLIC;
