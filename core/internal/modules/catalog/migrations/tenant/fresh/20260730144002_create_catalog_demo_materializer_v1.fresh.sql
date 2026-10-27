-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_demo_materialize_v1(
    p_package_id uuid, p_package_slug text, p_package_version text,
    p_release_id uuid, p_component_id uuid, p_owner_generation bigint,
    p_contract text, p_content_sha256 text, p_module_slug text,
    p_source_contract text, p_semantic_sha256 bytea,
    p_form_count integer, p_row_count integer,
    p_document_bytes bytea, p_occurred_at timestamptz
) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component catalog_catalog_components%ROWTYPE;
    v_payload catalog_demo_payloads%ROWTYPE;
BEGIN
    SELECT * INTO v_component
    FROM catalog_catalog_components AS component
    WHERE component.component_id = p_component_id
    FOR UPDATE;
    IF NOT FOUND THEN
        INSERT INTO catalog_catalog_components(
            component_id, component_kind, contract, content_sha256,
            owner_generation, writer_state, lifecycle_state, updated_at
        ) VALUES (
            p_component_id, 'DEMO_DATA', p_contract, p_content_sha256,
            p_owner_generation, 'v2', 'AVAILABLE', p_occurred_at
        );
    ELSIF ROW(
              v_component.component_kind, v_component.contract,
              v_component.owner_generation, v_component.writer_state
          ) IS DISTINCT FROM ROW(
              'DEMO_DATA'::text, p_contract,
              p_owner_generation, 'v2'::text
          ) OR v_component.lifecycle_state NOT IN (
              'AVAILABLE','STAGED','ACTIVE','DISABLED'
          ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog demo component authority differs';
    END IF;
    IF EXISTS (
        SELECT 1 FROM catalog_demo_payloads AS prior
        WHERE prior.component_id = p_component_id
          AND prior.package_id <> p_package_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='Catalog demo component package differs';
    END IF;
    SELECT * INTO v_payload
    FROM catalog_demo_payloads AS payload
    WHERE payload.package_id = p_package_id
      AND payload.package_version = p_package_version
      AND payload.component_id = p_component_id
    FOR UPDATE;
    IF NOT FOUND THEN
        INSERT INTO catalog_demo_payloads(
            component_id, package_id, package_slug, package_version,
            release_id, module_slug, source_contract, semantic_sha256,
            form_count, row_count, document_bytes, installed_at
        ) VALUES (
            p_component_id, p_package_id, p_package_slug, p_package_version,
            p_release_id, p_module_slug, p_source_contract, p_semantic_sha256,
            p_form_count, p_row_count, p_document_bytes, p_occurred_at
        );
    ELSIF ROW(
              v_payload.package_id, v_payload.package_slug,
              v_payload.package_version, v_payload.release_id,
              v_payload.module_slug, v_payload.source_contract,
              v_payload.semantic_sha256, v_payload.form_count,
              v_payload.row_count, v_payload.document_bytes
          ) IS DISTINCT FROM ROW(
              p_package_id, p_package_slug, p_package_version, p_release_id,
              p_module_slug, p_source_contract, p_semantic_sha256,
              p_form_count, p_row_count, p_document_bytes
          ) THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='Catalog demo semantic projection differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION catalog_demo_materialize_v1(
    uuid,text,text,uuid,uuid,bigint,text,text,text,text,bytea,
    integer,integer,bytea,timestamptz
) FROM PUBLIC;
