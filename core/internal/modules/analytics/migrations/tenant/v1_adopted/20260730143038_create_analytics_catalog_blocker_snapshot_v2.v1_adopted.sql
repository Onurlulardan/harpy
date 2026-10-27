-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_blocker_snapshot_v2(
    p_component_id uuid,p_reference_epoch bigint,p_generation bigint
)
RETURNS TABLE(
    blockers jsonb,zero_digest text,analytics_kind text,
    root_id uuid,head_version_id uuid,root_reference_epoch bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component_kind text;
    v_package_id uuid;
    v_package_version text;
    v_release_id uuid;
BEGIN
    IF pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_reference_epoch<1 OR p_generation<1 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog blocker snapshot input is invalid';
    END IF;
    SELECT projection.analytics_kind,projection.operational_resource_id,projection.published_version_id,component.component_kind,
        projection.package_id,projection.package_version,projection.release_id
    INTO analytics_kind,root_id,head_version_id,v_component_kind,v_package_id,v_package_version,v_release_id
    FROM analytics_catalog_components AS component
    JOIN analytics_catalog_component_heads AS head
      ON head.component_id=component.component_id
     AND head.content_sha256=component.content_sha256
    JOIN analytics_catalog_version_projections AS projection
      ON projection.receipt_id=head.projection_receipt_id
     AND projection.component_id=head.component_id
     AND projection.package_id=head.package_id
     AND projection.package_version=head.package_version
     AND projection.release_id=head.release_id
     AND projection.content_sha256=head.content_sha256
    JOIN analytics_catalog_definitions AS definition
      ON definition.component_id=projection.component_id
     AND definition.package_id=projection.package_id
     AND definition.package_version=projection.package_version
     AND definition.release_id=projection.release_id
     AND definition.content_sha256=projection.content_sha256
     AND definition.operational_resource_id=projection.operational_resource_id
     AND definition.published_version_id=projection.published_version_id
    WHERE component.component_id=p_component_id
      AND component.owner_generation=p_generation
      AND component.writer_state='v2'
      AND component.lifecycle_state='RETIRED'
      AND projection.operational_state='ready'
      AND projection.ready_at IS NOT NULL
    FOR SHARE OF component,head,projection,definition;
    IF NOT FOUND OR
       (analytics_kind='endpoint' AND v_component_kind<>'ENDPOINT') OR
       (analytics_kind='dashboard' AND v_component_kind<>'DASHBOARD') THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog blocker target differs';
    END IF;
    IF analytics_kind='endpoint' THEN
        SELECT root.reference_epoch INTO root_reference_epoch
        FROM analytics_saved_queries AS root
        JOIN analytics_saved_query_versions AS version_row
          ON version_row.query_id=root.id
         AND version_row.id=head_version_id
        WHERE root.id=root_id AND root.lifecycle_state='retired'
          AND version_row.version_state IN (
              'published','needs_revalidation','retired'
          )
        FOR SHARE OF root,version_row;
    ELSE
        SELECT root.reference_epoch INTO root_reference_epoch
        FROM analytics_dashboards AS root
        JOIN analytics_dashboard_versions AS version_row
          ON version_row.dashboard_id=root.id
         AND version_row.id=head_version_id
        WHERE root.id=root_id AND root.lifecycle_state='retired'
          AND version_row.version_state IN ('published','retired')
        FOR SHARE OF root,version_row;
    END IF;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog blocker root differs';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:legal-hold',6100
    ));
    WITH root_versions AS (
        SELECT version_row.id,'query'::text AS resource_kind,
            'query_version'::text AS target_kind
        FROM analytics_saved_query_versions AS version_row
        WHERE analytics_kind='endpoint' AND version_row.query_id=root_id
        UNION ALL
        SELECT version_row.id,'dashboard'::text,'dashboard_version'::text
        FROM analytics_dashboard_versions AS version_row
        WHERE analytics_kind='dashboard' AND version_row.dashboard_id=root_id
    ), publication_ids AS (
        SELECT release.id
        FROM analytics_publication_releases AS release
        JOIN root_versions AS version_row
          ON version_row.id=release.resource_version_id
         AND version_row.resource_kind=release.resource_kind
    ), raw_blockers AS (
        SELECT 'ANALYTICS_ACTIVE_INBOUND_REFERENCE'::text AS code,
            reference_row.reference_id::text AS identity
        FROM analytics_inbound_references AS reference_row
        WHERE reference_row.state='active' AND (
            EXISTS (
                SELECT 1 FROM root_versions AS version_row
                WHERE version_row.target_kind=reference_row.target_kind
                  AND version_row.id=reference_row.target_id
            ) OR (
                reference_row.target_kind='publication_release' AND
                reference_row.target_id IN (SELECT id FROM publication_ids)
            )
        )
        UNION ALL
        SELECT 'ANALYTICS_DASHBOARD_WIDGET_REFERENCE',
            widget.dashboard_version_id::text||':'||widget.widget_id::text
        FROM analytics_dashboard_widgets AS widget
        JOIN root_versions AS version_row
          ON version_row.resource_kind='query'
         AND version_row.id=widget.query_version_id
        WHERE NOT analytics_catalog_internal_widget_peer_v2(
            v_package_id,v_package_version,v_release_id,
            widget.dashboard_version_id
        )
        UNION ALL
        SELECT 'ANALYTICS_PUBLICATION_RELEASE_REFERENCE',id::text
        FROM publication_ids
        UNION ALL
        SELECT 'ANALYTICS_LEGAL_HOLD',hold.hold_id::text||':'||
            pg_catalog.encode(hold.reason_sha256,'hex')
        FROM analytics_legal_holds AS hold WHERE hold.state='active'
        UNION ALL
        SELECT blocker.code,blocker.digest
        FROM analytics_catalog_component_blockers AS blocker
        WHERE blocker.component_id=p_component_id
          AND blocker.active
        UNION ALL
        SELECT 'ANALYTICS_CATALOG_CANDIDATE',
            candidate.projection_receipt_id::text||':'||
            candidate.attempt_receipt_id::text
        FROM analytics_catalog_component_candidates AS candidate
        WHERE candidate.component_id=p_component_id
    ), grouped AS (
        SELECT code,pg_catalog.string_agg(
            identity,E'\\000' ORDER BY identity
        ) AS identities
        FROM raw_blockers GROUP BY code
    )
    SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
        'Owner','analytics','Code',code,'Digest',
        pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
            'analytics-catalog-blocker.v2'||E'\\000'||p_component_id::text||
            E'\\000'||code||E'\\000'||identities,'UTF8'
        )),'hex')
    ) ORDER BY code),'[]'::jsonb) INTO blockers
    FROM grouped;
    zero_digest:=pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        'analytics-catalog-blockers.v2'||E'\\000'||p_component_id::text||
        E'\\000'||p_reference_epoch::text||E'\\000'||analytics_kind||
        E'\\000'||v_package_id::text||E'\\000'||v_package_version||
        E'\\000'||v_release_id::text||E'\\000'||root_id::text||
        E'\\000'||head_version_id::text||
        E'\\000'||root_reference_epoch::text||E'\\000'||blockers::text,
        'UTF8'
    )),'hex');
    RETURN NEXT;
END;
$function$;
REVOKE ALL ON FUNCTION analytics_catalog_blocker_snapshot_v2(
    uuid,bigint,bigint
) FROM PUBLIC;
-- +goose StatementEnd
