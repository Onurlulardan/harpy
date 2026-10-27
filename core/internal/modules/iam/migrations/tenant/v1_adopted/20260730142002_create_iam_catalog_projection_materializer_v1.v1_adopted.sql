-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_materialize_v1(
    p_package_id uuid, p_package_slug text, p_package_version text,
    p_release_id uuid, p_component_id uuid, p_owner_generation bigint,
    p_contract text, p_content_sha256 text, p_component_kind text,
    p_module_slug text, p_semantic_sha256 bytea,
    p_document_bytes bytea, p_document jsonb, p_occurred_at timestamptz
) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component iam_catalog_components%ROWTYPE;
    v_materialization iam_catalog_materializations%ROWTYPE;
BEGIN
    SELECT * INTO v_component
    FROM iam_catalog_components AS component
    WHERE component.component_id = p_component_id
    FOR UPDATE;
    IF NOT FOUND THEN
        INSERT INTO iam_catalog_components(
            component_id, component_kind, contract, content_sha256,
            owner_generation, writer_state, lifecycle_state, updated_at
        ) VALUES (
            p_component_id, p_component_kind, p_contract, p_content_sha256,
            p_owner_generation, 'v2', 'AVAILABLE', p_occurred_at
        );
    ELSIF ROW(v_component.component_kind, v_component.contract,
              v_component.content_sha256, v_component.owner_generation,
              v_component.writer_state)
          IS DISTINCT FROM
          ROW(p_component_kind, p_contract, p_content_sha256,
              p_owner_generation, 'v2'::text) OR
          v_component.lifecycle_state NOT IN ('AVAILABLE','STAGED') THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM catalog component authority differs';
    END IF;
    SELECT * INTO v_materialization
    FROM iam_catalog_materializations AS materialization
    WHERE materialization.component_id = p_component_id
    FOR UPDATE;
    IF NOT FOUND THEN
        INSERT INTO iam_catalog_materializations(
            component_id, module_slug, content_sha256, semantic_sha256,
            document_bytes, package_id, package_slug,
            package_version, release_id, installed_at
        ) VALUES (
            p_component_id, p_module_slug, p_content_sha256,
            p_semantic_sha256, p_document_bytes, p_package_id,
            p_package_slug, p_package_version, p_release_id, p_occurred_at
        );
        INSERT INTO iam_catalog_resource_projections(
            component_id, resource_slug, resource_kind, name, description
        )
        SELECT p_component_id, resource.slug, resource.kind,
            resource.name, resource.description
        FROM jsonb_to_recordset(p_document->'resources')
            AS resource(slug text, kind text, name text, description text);
        INSERT INTO iam_catalog_grant_projections(
            component_id, resource_slug, role_name, role_description, actions
        )
        SELECT p_component_id, permission.resource_slug, permission.role_name,
            permission.role_description,
            ARRAY(SELECT jsonb_array_elements_text(permission.actions))
        FROM jsonb_to_recordset(p_document->'grants')
            AS permission(
                resource_slug text, role_name text,
                role_description text, actions jsonb
            );
    ELSIF ROW(v_materialization.module_slug,
              v_materialization.content_sha256,
              v_materialization.semantic_sha256,
              v_materialization.document_bytes)
          IS DISTINCT FROM
          ROW(p_module_slug, p_content_sha256,
              p_semantic_sha256, p_document_bytes) THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='IAM catalog semantic projection differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_catalog_materialize_v1(
    uuid,text,text,uuid,uuid,bigint,text,text,text,text,
    bytea,bytea,jsonb,timestamptz
) FROM PUBLIC;
