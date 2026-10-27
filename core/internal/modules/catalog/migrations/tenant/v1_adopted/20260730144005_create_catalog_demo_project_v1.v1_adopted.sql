-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_demo_project_v1(
    p_scope_schema name, p_receipt_id uuid,
    p_package_id uuid, p_package_slug text, p_package_version text,
    p_release_id uuid, p_component_id uuid, p_owner_generation bigint,
    p_contract text, p_content_sha256 text, p_module_slug text,
    p_source_contract text, p_semantic_sha256 bytea,
    p_form_count integer, p_row_count integer, p_document_bytes bytea,
    p_idempotency_key text, p_request_sha256 bytea,
    p_occurred_at timestamptz
)
RETURNS TABLE(
    receipt_id uuid, package_id uuid, package_slug text,
    package_version text, release_id uuid, component_id uuid,
    owner_generation bigint, contract text, content_sha256 text,
    module_slug text, source_contract text, semantic_sha256 bytea,
    form_count integer, row_count integer, document_bytes bytea,
    idempotency_key text, request_sha256 bytea, lifecycle_state text,
    ready boolean, retry_at timestamptz, completed_at timestamptz,
    created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_document jsonb;
    v_receipt catalog_demo_provisions%ROWTYPE;
    v_payload catalog_demo_payloads%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       p_package_id IS NULL OR
       p_package_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       p_package_version IS NULL OR
       p_package_version !~
        '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$' OR
       p_owner_generation < 1 OR
       length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_idempotency_key !~ '^[A-Za-z0-9_.:-]+$' OR
       octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) OR
       date_trunc('microseconds', p_occurred_at) <> p_occurred_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Catalog demo projection request is invalid';
    END IF;
    v_document := catalog_demo_project_document_v1(
        p_package_slug, p_component_id, p_contract, p_content_sha256,
        p_module_slug, p_source_contract, p_semantic_sha256,
        p_form_count, p_row_count, p_document_bytes
    );
    PERFORM catalog_demo_assert_current_authority_v1(
        p_scope_schema, p_owner_generation
    );
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'catalog:demo-project:' || p_component_id::text, 6104
    ));
    SELECT * INTO v_receipt
    FROM catalog_demo_provisions AS provision
    WHERE provision.package_id = p_package_id
      AND provision.package_version = p_package_version
      AND provision.component_id = p_component_id
    FOR UPDATE;
    IF FOUND THEN
        SELECT * INTO STRICT v_payload
        FROM catalog_demo_payloads AS payload
        WHERE payload.package_id = p_package_id
          AND payload.package_version = p_package_version
          AND payload.component_id = p_component_id
        FOR SHARE;
        IF ROW(
              v_receipt.package_id, v_receipt.package_slug,
              v_receipt.package_version, v_receipt.release_id,
              v_receipt.owner_generation, v_receipt.contract,
              v_receipt.content_sha256, v_receipt.idempotency_key,
              v_receipt.request_sha256
           ) IS DISTINCT FROM ROW(
              p_package_id, p_package_slug, p_package_version, p_release_id,
              p_owner_generation, p_contract, p_content_sha256,
              p_idempotency_key, p_request_sha256
           ) OR ROW(
              v_payload.module_slug, v_payload.source_contract,
              v_payload.semantic_sha256, v_payload.form_count,
              v_payload.row_count, v_payload.document_bytes
           ) IS DISTINCT FROM ROW(
              p_module_slug, p_source_contract, p_semantic_sha256,
              p_form_count, p_row_count, p_document_bytes
           ) THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='Catalog demo projection replay differs';
        END IF;
        PERFORM catalog_demo_reconcile_v1(
            v_receipt.receipt_id, p_occurred_at
        );
        RETURN QUERY
        SELECT * FROM catalog_demo_readback_v1(v_receipt.receipt_id);
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Catalog demo replay readback differs';
        END IF;
        RETURN;
    END IF;
    PERFORM catalog_demo_materialize_v1(
        p_package_id, p_package_slug, p_package_version,
        p_release_id, p_component_id, p_owner_generation,
        p_contract, p_content_sha256, p_module_slug,
        p_source_contract, p_semantic_sha256,
        p_form_count, p_row_count, p_document_bytes, p_occurred_at
    );
    INSERT INTO catalog_demo_provisions(
        receipt_id, package_id, package_slug, package_version,
        release_id, component_id, owner_generation, contract,
        content_sha256, idempotency_key, request_sha256,
        ready, retry_at, completed_at, created_at, updated_at
    ) VALUES (
        p_receipt_id, p_package_id, p_package_slug, p_package_version,
        p_release_id, p_component_id, p_owner_generation, p_contract,
        p_content_sha256, p_idempotency_key, p_request_sha256,
        false, p_occurred_at + interval '30 seconds',
        p_occurred_at, p_occurred_at, p_occurred_at
    );
    RETURN QUERY SELECT * FROM catalog_demo_readback_v1(p_receipt_id);
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog demo projection readback differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION catalog_demo_project_v1(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,
    bytea,integer,integer,bytea,text,bytea,timestamptz
) FROM PUBLIC;
