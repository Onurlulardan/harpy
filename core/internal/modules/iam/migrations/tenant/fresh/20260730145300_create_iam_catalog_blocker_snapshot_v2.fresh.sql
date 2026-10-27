-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_blocker_snapshot_v2(
    p_component_id uuid,p_reference_epoch bigint,p_generation bigint
)
RETURNS TABLE(blockers jsonb,zero_digest text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component iam_catalog_components%ROWTYPE;
    v_head iam_catalog_component_heads%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_reference_epoch IS NULL OR p_reference_epoch<1 OR
       p_generation IS NULL OR p_generation<1 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM Catalog blocker snapshot input is invalid';
    END IF;
    SELECT * INTO v_component
    FROM iam_catalog_components AS component
    WHERE component.component_id=p_component_id FOR SHARE;
    SELECT * INTO v_head
    FROM iam_catalog_component_heads AS head
    WHERE head.component_id=p_component_id FOR SHARE;
    IF v_component.component_id IS NULL OR v_head.component_id IS NULL OR
       v_component.component_kind<>'IAM_TEMPLATE' OR
       v_component.owner_generation<>p_generation OR
       v_component.writer_state<>'v2' OR
       v_component.lifecycle_state<>'RETIRED' OR
       v_head.owner_generation<>p_generation OR
       v_head.content_sha256<>v_component.content_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM Catalog blocker snapshot authority differs';
    END IF;
    RETURN QUERY
    WITH mapped_resources AS (
        SELECT DISTINCT mapped.resource_id
        FROM iam_catalog_version_operational_resources AS mapped
        WHERE mapped.component_id=p_component_id
    ),mapped_permissions AS (
        SELECT DISTINCT mapped.permission_id
        FROM iam_catalog_version_operational_grants AS mapped
        WHERE mapped.component_id=p_component_id
    ),mapped_bindings AS (
        SELECT DISTINCT mapped.binding_id
        FROM iam_catalog_version_operational_grants AS mapped
        WHERE mapped.component_id=p_component_id
    ),external_permission_references AS (
        SELECT 'role:'||binding.id::text AS reference
        FROM iam_role_permission_bindings AS binding
        JOIN mapped_permissions AS permission
          ON permission.permission_id=binding.permission_id
        LEFT JOIN mapped_bindings AS mapped ON mapped.binding_id=binding.id
        WHERE binding.state='active' AND mapped.binding_id IS NULL
        UNION ALL
        SELECT 'user:'||binding.id::text
        FROM iam_user_permission_bindings AS binding
        JOIN mapped_permissions AS permission
          ON permission.permission_id=binding.permission_id
        WHERE binding.state='active'
        UNION ALL
        SELECT 'organization:'||binding.id::text
        FROM iam_organization_permission_bindings AS binding
        JOIN mapped_permissions AS permission
          ON permission.permission_id=binding.permission_id
        WHERE binding.state='active'
        UNION ALL
        SELECT 'implication:'||implication.id::text
        FROM iam_permission_implications AS implication
        WHERE implication.state='active' AND (
            EXISTS (SELECT 1 FROM mapped_permissions AS mapped
                WHERE mapped.permission_id=implication.source_permission_id) OR
            EXISTS (SELECT 1 FROM mapped_permissions AS mapped
                WHERE mapped.permission_id=implication.implied_permission_id)
        )
        UNION ALL
        SELECT 'assignment:'||member.group_id::text||':'||
            member.group_revision::text||':'||member.ordinal::text
        FROM iam_rbac_assignment_members AS member
        JOIN mapped_permissions AS permission
          ON permission.permission_id=member.permission_id
        WHERE member.state='active'
        UNION ALL
        SELECT 'request:'||request.id::text
        FROM iam_permission_requests AS request
        JOIN mapped_permissions AS permission
          ON permission.permission_id=request.permission_id
        WHERE request.state IN ('PENDING','IN_REVIEW')
    ),facts AS (
        SELECT blocker.code,blocker.digest
        FROM iam_catalog_component_blockers AS blocker
        WHERE blocker.component_id=p_component_id AND blocker.active
        UNION ALL
        SELECT 'IAM_CATALOG_CANDIDATE_ACTIVE',
            pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
                pg_catalog.string_agg(
                    candidate.provision_receipt_id::text,',' ORDER BY
                    candidate.provision_receipt_id
                ),'UTF8')),'hex')
        FROM iam_catalog_component_candidates AS candidate
        WHERE candidate.component_id=p_component_id
        HAVING pg_catalog.count(*)>0
        UNION ALL
        SELECT 'IAM_CATALOG_GRANT_ACTIVE',
            pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
                pg_catalog.string_agg(
                    binding.id::text,',' ORDER BY binding.id
                ),'UTF8')),'hex')
        FROM iam_role_permission_bindings AS binding
        JOIN mapped_bindings AS mapped ON mapped.binding_id=binding.id
        WHERE binding.state='active'
        HAVING pg_catalog.count(*)>0
        UNION ALL
        SELECT 'IAM_CATALOG_EXTERNAL_PERMISSION',
            pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
                pg_catalog.string_agg(
                    permission.id::text,',' ORDER BY permission.id
                ),'UTF8')),'hex')
        FROM iam_permissions AS permission
        JOIN mapped_resources AS resource
          ON resource.resource_id=permission.resource_id
        LEFT JOIN mapped_permissions AS mapped
          ON mapped.permission_id=permission.id
        WHERE permission.state='active' AND mapped.permission_id IS NULL
        HAVING pg_catalog.count(*)>0
        UNION ALL
        SELECT 'IAM_CATALOG_EXTERNAL_PERMISSION_REFERENCE',
            pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
                pg_catalog.string_agg(
                    reference.reference,',' ORDER BY reference.reference
                ),'UTF8')),'hex')
        FROM external_permission_references AS reference
        HAVING pg_catalog.count(*)>0
    ),document AS (
        SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
            'Owner','iam','Code',fact.code,'Digest',fact.digest
        ) ORDER BY fact.code,fact.digest),'[]'::jsonb) AS blockers
        FROM facts AS fact
    )
    SELECT document.blockers,
        pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
            'iam-catalog-blockers.v2'||E'\\000'||p_component_id::text||
            E'\\000'||p_reference_epoch::text||E'\\000'||
            v_head.package_id::text||E'\\000'||v_head.package_version||
            E'\\000'||v_head.release_id::text||E'\\000'||
            v_head.content_sha256||E'\\000'||document.blockers::text,
            'UTF8'
        )),'hex')
    FROM document;
END;
$function$;
REVOKE ALL ON FUNCTION iam_catalog_blocker_snapshot_v2(
    uuid,bigint,bigint
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION iam_catalog_blocker_snapshot_v2(
    uuid,bigint,bigint
) FROM forgecrud_runtime;
-- +goose StatementEnd
