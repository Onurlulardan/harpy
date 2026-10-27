-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_materialize_v2(
    p_package_id uuid,p_package_slug text,p_package_version text,
    p_release_id uuid,p_component_id uuid,p_owner_generation bigint,
    p_contract text,p_content_sha256 text,p_component_kind text,
    p_module_slug text,p_semantic_identity text,p_semantic_sha256 bytea,
    p_item_count integer,p_auxiliary_count integer,
    p_document_bytes bytea,p_document jsonb,p_occurred_at timestamptz
) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component iam_catalog_components%ROWTYPE;
    v_evidence iam_catalog_version_evidence%ROWTYPE;
    v_has_head boolean;
BEGIN
    SELECT * INTO v_component
    FROM iam_catalog_components AS component
    WHERE component.component_id=p_component_id
    FOR UPDATE;
    IF NOT FOUND THEN
        INSERT INTO iam_catalog_components(
            component_id,component_kind,contract,content_sha256,
            owner_generation,writer_state,lifecycle_state,updated_at
        ) VALUES (
            p_component_id,p_component_kind,p_contract,p_content_sha256,
            p_owner_generation,'v2','AVAILABLE',p_occurred_at
        ) RETURNING * INTO v_component;
    ELSIF ROW(
              v_component.component_kind,v_component.contract,
              v_component.owner_generation,v_component.writer_state
          ) IS DISTINCT FROM ROW(
              p_component_kind,p_contract,p_owner_generation,'v2'::text
          ) OR v_component.lifecycle_state NOT IN (
              'AVAILABLE','STAGED','ACTIVE','DISABLED'
          ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM catalog component authority differs';
    END IF;
    SELECT EXISTS(
        SELECT 1 FROM iam_catalog_component_heads AS head
        WHERE head.component_id=p_component_id
    ) INTO v_has_head;
    IF v_has_head AND NOT EXISTS (
        SELECT 1 FROM iam_catalog_component_heads AS head
        WHERE head.component_id=p_component_id
          AND head.owner_generation=v_component.owner_generation
          AND head.content_sha256=v_component.content_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM catalog active head differs';
    ELSIF NOT v_has_head AND v_component.lifecycle_state='STAGED'
          AND v_component.content_sha256<>p_content_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM catalog staged component differs';
    ELSIF NOT v_has_head AND v_component.lifecycle_state='AVAILABLE'
          AND v_component.content_sha256<>p_content_sha256 THEN
        UPDATE iam_catalog_components
        SET content_sha256=p_content_sha256,updated_at=p_occurred_at
        WHERE component_id=p_component_id;
    END IF;
    IF EXISTS (
        SELECT 1 FROM iam_catalog_version_evidence AS prior
        WHERE prior.component_id=p_component_id
          AND prior.package_id<>p_package_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='IAM catalog component package differs';
    END IF;
    SELECT * INTO v_evidence
    FROM iam_catalog_version_evidence AS evidence
    WHERE evidence.package_id=p_package_id
      AND evidence.package_version=p_package_version
      AND evidence.component_id=p_component_id
    FOR SHARE;
    IF NOT FOUND THEN
        INSERT INTO iam_catalog_version_evidence(
            component_id,package_id,package_slug,package_version,
            release_id,owner_generation,contract,content_sha256,
            component_kind,module_slug,semantic_identity,semantic_sha256,
            item_count,auxiliary_count,document_bytes,installed_at
        ) VALUES (
            p_component_id,p_package_id,p_package_slug,p_package_version,
            p_release_id,p_owner_generation,p_contract,p_content_sha256,
            p_component_kind,p_module_slug,p_semantic_identity,
            p_semantic_sha256,p_item_count,p_auxiliary_count,
            p_document_bytes,p_occurred_at
        );
        INSERT INTO iam_catalog_version_resources(
            component_id,package_id,package_version,release_id,
            owner_generation,content_sha256,resource_slug,
            resource_kind,name,description
        )
        SELECT p_component_id,p_package_id,p_package_version,p_release_id,
            p_owner_generation,p_content_sha256,resource.slug,
            resource.kind,resource.name,resource.description
        FROM pg_catalog.jsonb_to_recordset(p_document->'resources')
            AS resource(slug text,kind text,name text,description text);
        INSERT INTO iam_catalog_version_grants(
            component_id,package_id,package_version,release_id,
            owner_generation,content_sha256,resource_slug,
            role_name,role_description,actions
        )
        SELECT p_component_id,p_package_id,p_package_version,p_release_id,
            p_owner_generation,p_content_sha256,grant_row.resource_slug,
            grant_row.role_name,grant_row.role_description,
            ARRAY(SELECT pg_catalog.jsonb_array_elements_text(grant_row.actions))
        FROM pg_catalog.jsonb_to_recordset(p_document->'grants')
            AS grant_row(
                resource_slug text,role_name text,
                role_description text,actions jsonb
            );
    ELSIF ROW(
              v_evidence.package_slug,v_evidence.release_id,
              v_evidence.owner_generation,v_evidence.contract,
              v_evidence.content_sha256,v_evidence.component_kind,
              v_evidence.module_slug,v_evidence.semantic_identity,
              v_evidence.semantic_sha256,v_evidence.item_count,
              v_evidence.auxiliary_count,v_evidence.document_bytes
          ) IS DISTINCT FROM ROW(
              p_package_slug,p_release_id,p_owner_generation,p_contract,
              p_content_sha256,p_component_kind,p_module_slug,
              p_semantic_identity,p_semantic_sha256,p_item_count,
              p_auxiliary_count,p_document_bytes
          ) THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='IAM catalog version evidence differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_catalog_materialize_v2(
    uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,jsonb,timestamptz
) FROM PUBLIC;
