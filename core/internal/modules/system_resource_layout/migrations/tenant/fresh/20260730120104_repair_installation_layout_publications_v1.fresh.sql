-- +goose Up
-- Migration stream version: 20260730120104.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_repair_installation_defaults_v1(
    p_successor_contract text
)
RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_current system_resource_layout_publications%ROWTYPE;
    v_revision_id uuid;
    v_created_at timestamptz;
    v_definition text;
    v_presentation text;
    v_definition_sha bytea;
    v_presentation_sha bytea;
    v_version bigint;
    v_changed bigint;
BEGIN
    IF p_successor_contract IS DISTINCT FROM
       'curated-installation-default-successor.v1' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'system resource successor contract is invalid';
    END IF;
    IF NOT system_resource_layout_installation_baseline_evidence_valid_v1(
        'installation-baseline-evidence.v1'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'system resource installation baseline evidence is invalid';
    END IF;
    v_changed := 0;
    FOR v_current IN
        SELECT publication.*
        FROM system_resource_layout_publications AS publication
        JOIN system_resource_layout_revisions AS revision
          ON revision.id = publication.revision_id
         AND revision.resource_kind = publication.resource_kind
        JOIN system_resource_layout_installation_baselines AS baseline
          ON publication.revision_id = CASE publication.resource_kind
             WHEN 'user' THEN baseline.user_revision_id
             WHEN 'organization' THEN baseline.organization_revision_id
             WHEN 'role' THEN baseline.role_revision_id END
        WHERE publication.generation = 1
          AND publication.catalog_contract =
              'forms-presentation-field-catalog.v1'
          AND revision.state = 'published'
        ORDER BY publication.resource_kind
    LOOP
        SELECT greatest(v_current.updated_at, revision.created_at) +
               INTERVAL '1 microsecond'
        INTO v_created_at
        FROM system_resource_layout_revisions AS revision
        WHERE revision.id = v_current.revision_id;
        v_revision_id := system_resource_layout_deterministic_uuid_v7_v1(
            v_current.revision_id::text, p_successor_contract, v_created_at
        );
        v_definition := system_resource_layout_curated_definition_v1(
            v_current.resource_kind, v_current.catalog_json
        );
        v_definition_sha := pg_catalog.sha256(
            pg_catalog.convert_to(v_definition, 'UTF8')
        );
        v_presentation := system_resource_layout_compile_curated_v1(
            v_revision_id, v_created_at, v_definition,
            v_current.catalog_json
        );
        v_presentation_sha := pg_catalog.sha256(
            pg_catalog.convert_to(v_presentation, 'UTF8')
        );
        SELECT COALESCE(pg_catalog.max(revision.version), 0) + 1
        INTO v_version
        FROM system_resource_layout_revisions AS revision
        WHERE revision.resource_kind = v_current.resource_kind;
        INSERT INTO system_resource_layout_revisions (
            id, resource_kind, version, state, definition_json,
            definition_sha256, created_at, updated_at, published_at,
            row_revision
        ) VALUES (
            v_revision_id, v_current.resource_kind, v_version, 'draft',
            v_definition, v_definition_sha, v_created_at, v_created_at,
            NULL, 1
        );
        UPDATE system_resource_layout_revisions
        SET state = 'published', published_at = v_created_at,
            updated_at = v_created_at, row_revision = 2
        WHERE id = v_revision_id AND state = 'draft' AND row_revision = 1;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'system resource curated successor publish conflicted';
        END IF;
        UPDATE system_resource_layout_publications
        SET revision_id = v_revision_id,
            generation = v_current.generation + 1,
            updated_at = v_created_at,
            source_definition_sha256 = v_definition_sha,
            compiler_version = 'system-resource-presentation-compiler.v2',
            catalog_contract = v_current.catalog_contract,
            catalog_json = v_current.catalog_json,
            catalog_sha256 = v_current.catalog_sha256,
            presentation_contract = 'forms-presentation.v1',
            presentation_json = v_presentation,
            presentation_sha256 = v_presentation_sha
        WHERE resource_kind = v_current.resource_kind
          AND revision_id = v_current.revision_id
          AND generation = v_current.generation;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'system resource curated publication conflicted';
        END IF;
        v_changed := v_changed + 1;
    END LOOP;
    RETURN v_changed;
END;
$function$;

SELECT system_resource_layout_repair_installation_defaults_v1(
    'curated-installation-default-successor.v1'
);
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_repair_installation_defaults_v1(text)
    FROM PUBLIC;
