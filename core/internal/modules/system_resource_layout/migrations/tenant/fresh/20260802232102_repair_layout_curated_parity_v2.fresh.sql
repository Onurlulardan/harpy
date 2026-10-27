-- +goose Up
-- Migration stream version: 20260802232102.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_repair_curated_defaults_v3(
    p_successor_contract text
)
RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_current system_resource_layout_publications%ROWTYPE;
    v_revision system_resource_layout_revisions%ROWTYPE;
    v_baseline_revision_id uuid;
    v_revision_id uuid; v_created_at timestamptz; v_catalog text;
    v_definition text; v_presentation text; v_definition_sha bytea;
    v_presentation_sha bytea; v_catalog_sha bytea;
    v_next_version bigint; v_changed bigint:=0;
BEGIN
    IF p_successor_contract IS DISTINCT FROM
       'curated-system-resource-parity-successor.v2' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='system resource curated parity contract is invalid';
    END IF;
    IF (SELECT pg_catalog.count(*) FROM
        system_resource_layout_installation_baselines)=0 THEN RETURN 0; END IF;
    FOR v_current IN
        SELECT publication.* FROM system_resource_layout_publications AS publication
        WHERE publication.generation=2 ORDER BY publication.resource_kind
    LOOP
        SELECT revision.* INTO STRICT v_revision
          FROM system_resource_layout_revisions AS revision
         WHERE revision.id=v_current.revision_id;
        IF NOT system_resource_layout_curated_predecessor_valid_v1(
            v_current.resource_kind) THEN
            SELECT CASE v_current.resource_kind
                WHEN 'user' THEN baseline.user_revision_id
                WHEN 'organization' THEN baseline.organization_revision_id
                ELSE baseline.role_revision_id END
              INTO STRICT v_baseline_revision_id
              FROM system_resource_layout_installation_baselines AS baseline;
            IF v_revision.id=system_resource_layout_deterministic_uuid_v7_v1(
                v_baseline_revision_id::text,
                'curated-installation-default-successor.v1',
                v_revision.created_at) THEN
                RAISE EXCEPTION USING ERRCODE='23514',
                    MESSAGE='system resource curated predecessor is invalid';
            END IF;
            CONTINUE;
        END IF;
        v_created_at:=greatest(
            v_current.updated_at,v_revision.created_at)+INTERVAL '1 microsecond';
        v_revision_id:=system_resource_layout_deterministic_uuid_v7_v1(
            v_current.revision_id::text,p_successor_contract,v_created_at);
        v_catalog:=system_resource_layout_curated_successor_catalog_v1(
            v_current.resource_kind,v_current.catalog_json);
        v_catalog_sha:=pg_catalog.sha256(pg_catalog.convert_to(v_catalog,'UTF8'));
        v_definition:=system_resource_layout_curated_definition_v2(
            v_current.resource_kind,v_catalog);
        v_definition_sha:=pg_catalog.sha256(
            pg_catalog.convert_to(v_definition,'UTF8'));
        v_presentation:=system_resource_layout_compile_curated_v1(
            v_revision_id,v_created_at,v_definition,v_catalog);
        v_presentation_sha:=pg_catalog.sha256(
            pg_catalog.convert_to(v_presentation,'UTF8'));
        SELECT COALESCE(pg_catalog.max(revision.version),0)+1
          INTO v_next_version FROM system_resource_layout_revisions AS revision
         WHERE revision.resource_kind=v_current.resource_kind;
        INSERT INTO system_resource_layout_revisions
            (id,resource_kind,version,state,definition_json,definition_sha256,
             created_at,updated_at,published_at,row_revision)
        VALUES(v_revision_id,v_current.resource_kind,v_next_version,'draft',
            v_definition,v_definition_sha,v_created_at,v_created_at,NULL,1);
        UPDATE system_resource_layout_revisions AS revision SET state='published',
            published_at=v_created_at,updated_at=v_created_at,row_revision=2
         WHERE revision.id=v_revision_id AND revision.state='draft'
           AND revision.row_revision=1;
        IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='system resource curated parity revision conflicted'; END IF;
        UPDATE system_resource_layout_publications AS publication
           SET revision_id=v_revision_id,generation=v_current.generation+1,
               updated_at=v_created_at,source_definition_sha256=v_definition_sha,
               compiler_version='system-resource-presentation-compiler.v2',
               catalog_json=v_catalog,catalog_sha256=v_catalog_sha,
               presentation_json=v_presentation,
               presentation_sha256=v_presentation_sha
         WHERE publication.resource_kind=v_current.resource_kind
           AND publication.revision_id=v_current.revision_id
           AND publication.generation=v_current.generation;
        IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='system resource curated parity publication conflicted'; END IF;
        v_changed:=v_changed+1;
    END LOOP;
    RETURN v_changed;
END;
$function$;
SELECT system_resource_layout_repair_curated_defaults_v3(
    'curated-system-resource-parity-successor.v2');
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_repair_curated_defaults_v3(text)
    FROM PUBLIC;
