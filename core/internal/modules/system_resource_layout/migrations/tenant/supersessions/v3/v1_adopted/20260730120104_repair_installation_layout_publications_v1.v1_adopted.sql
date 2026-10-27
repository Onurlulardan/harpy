-- +goose Up
-- Migration stream version: 20260730120104.
-- Exact v3 supersession: skip schemas without an installation baseline.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_installation_catalog_valid_v2(
    p_resource_kind text, p_catalog_json text)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_catalog jsonb; v_prefix text; v_static_count integer; v_types jsonb;
    v_editable jsonb := '["user:002","user:006","user:007","role:002","role:003","role:004","role:005","role:006","role:009","organization:002","organization:003","organization:005"]';
    v_unique bigint; v_invalid bigint; v_static bigint; v_custom bigint;
BEGIN
    IF p_resource_kind NOT IN ('user','organization','role') THEN RETURN false; END IF;
    BEGIN
        v_catalog := p_catalog_json::jsonb;
    EXCEPTION WHEN OTHERS THEN RETURN false;
    END;
    v_prefix := CASE p_resource_kind WHEN 'user' THEN 'a1000000-' WHEN 'role' THEN 'b1000000-' ELSE 'c1000000-' END;
    v_static_count := CASE p_resource_kind WHEN 'role' THEN 14 ELSE 11 END;
    v_types := CASE p_resource_kind
        WHEN 'user' THEN '["uuid","text","text","enum","enum","uuid_list","uuid_list","uuid","integer","timestamp","timestamp"]'
        WHEN 'role' THEN '["uuid","enum","uuid","relation","text","text","enum","boolean","boolean","uuid_list","uuid","integer","timestamp","timestamp"]'
        ELSE '["uuid","text","text","enum","relation","uuid_list","uuid_list","uuid","integer","timestamp","timestamp"]'
        END;
    IF v_catalog IS NULL OR v_catalog->>'contract' IS DISTINCT FROM
       'forms-presentation-field-catalog.v1'
       OR pg_catalog.jsonb_typeof(v_catalog->'fields') IS DISTINCT FROM 'array'
       OR (SELECT pg_catalog.count(*) FROM
           pg_catalog.jsonb_object_keys(v_catalog)) <> 2
       OR pg_catalog.jsonb_array_length(v_catalog->'fields') <>
          v_static_count + 1 THEN RETURN false; END IF;
    WITH expected AS (
        SELECT v_prefix || '0000-4000-8000-' ||
               pg_catalog.lpad(pg_catalog.to_hex(number),12,'0') AS id,
               v_types->>(number-1) AS type_id
        FROM pg_catalog.generate_series(1,v_static_count) AS series(number)
    ), actual AS (
        SELECT field.value, field.value->>'id' AS id, expected.type_id
        FROM pg_catalog.jsonb_array_elements(v_catalog->'fields') AS field(value)
        LEFT JOIN expected ON expected.id = field.value->>'id'
    )
    SELECT pg_catalog.count(DISTINCT id),
           pg_catalog.count(*) FILTER (WHERE id IS NULL
             OR (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(value))<>4
             OR type_id IS NULL AND (
                id LIKE v_prefix||'%'
                OR id !~ '^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
                OR value->>'type_id' IS DISTINCT FROM 'text'
                OR value->>'value_profile' IS DISTINCT FROM 'text_nfc_v1'
                OR value->>'mode' IS DISTINCT FROM 'editable')
             OR type_id IS NOT NULL AND (
                value->>'type_id' IS DISTINCT FROM type_id
                OR value->>'value_profile' IS DISTINCT FROM CASE type_id
                   WHEN 'text' THEN 'text_nfc_v1'
                   WHEN 'uuid' THEN 'uuid_rfc4122_lower_v1'
                   WHEN 'enum' THEN 'enum_code_v1'
                   WHEN 'uuid_list' THEN 'uuid_list_v1'
                   WHEN 'integer' THEN 'integer_base10_v1'
                   WHEN 'timestamp' THEN 'timestamp_utc_microsecond_v1'
                   WHEN 'relation' THEN 'relation_uuid_v1'
                   WHEN 'boolean' THEN 'boolean_v1' END
                OR value->>'mode' IS DISTINCT FROM CASE
                   WHEN v_editable ? (p_resource_kind||':'||
                     pg_catalog.right(id,3)) THEN 'editable'
                   ELSE 'read_only' END)),
           pg_catalog.count(*) FILTER (WHERE type_id IS NOT NULL),
           pg_catalog.count(*) FILTER (WHERE type_id IS NULL)
    INTO v_unique,v_invalid,v_static,v_custom FROM actual;
    RETURN v_unique = v_static_count + 1 AND v_invalid = 0
       AND v_static = v_static_count AND v_custom = 1;
END;
$function$;
CREATE FUNCTION system_resource_layout_installation_baseline_evidence_valid_v2(
    p_contract text)
RETURNS boolean
LANGUAGE sql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
SELECT p_contract IS NOT DISTINCT FROM 'installation-baseline-evidence.v1'
AND (SELECT pg_catalog.count(*) FROM system_resource_layout_installation_baselines)=1 AND (SELECT pg_catalog.count(*) FROM system_resource_layout_publications)=3
AND NOT EXISTS (
 SELECT 1 FROM system_resource_layout_installation_baselines baseline
 CROSS JOIN (VALUES('user'),('organization'),('role')) expected(kind)
 LEFT JOIN system_resource_layout_publications publication
   ON publication.resource_kind=expected.kind
 LEFT JOIN system_resource_layout_revisions revision
   ON revision.id=CASE expected.kind WHEN 'user' THEN baseline.user_revision_id
      WHEN 'organization' THEN baseline.organization_revision_id
      ELSE baseline.role_revision_id END
 WHERE publication.resource_kind IS NULL OR publication.generation=1 AND (
       publication.revision_id IS DISTINCT FROM revision.id
    OR revision.resource_kind IS DISTINCT FROM expected.kind
    OR revision.state IS DISTINCT FROM 'published'
    OR publication.catalog_contract IS DISTINCT FROM 'forms-presentation-field-catalog.v1'
    OR publication.presentation_contract IS DISTINCT FROM 'forms-presentation.v1'
    OR publication.compiler_version IS DISTINCT FROM 'system-resource-presentation-compiler.v2'
    OR revision.definition_json::jsonb->>'contract' IS DISTINCT FROM
       'harpy/system-resource-layout/v2'
    OR publication.catalog_sha256 IS DISTINCT FROM pg_catalog.sha256(
       pg_catalog.convert_to(publication.catalog_json,'UTF8'))
    OR publication.presentation_sha256 IS DISTINCT FROM pg_catalog.sha256(
       pg_catalog.convert_to(publication.presentation_json,'UTF8'))
    OR revision.definition_sha256 IS DISTINCT FROM pg_catalog.sha256(
       pg_catalog.convert_to(revision.definition_json,'UTF8'))
    OR publication.source_definition_sha256 IS DISTINCT FROM revision.definition_sha256
    OR NOT system_resource_layout_installation_catalog_valid_v2(
       expected.kind,publication.catalog_json))
) AND NOT EXISTS (
 SELECT 1 FROM system_resource_layout_installation_baselines baseline
 JOIN system_resource_layout_publications users ON users.resource_kind='user'
 JOIN system_resource_layout_publications organizations ON organizations.resource_kind='organization'
 JOIN system_resource_layout_publications roles ON roles.resource_kind='role'
 WHERE users.generation=1 AND organizations.generation=1 AND roles.generation=1
 AND baseline.baseline_sha256 IS DISTINCT FROM pg_catalog.sha256(
   pg_catalog.uuid_send(baseline.operation_id)||pg_catalog.uuid_send(baseline.tenant_id)||
   pg_catalog.uuid_send(baseline.user_revision_id)||
   users.catalog_sha256||users.presentation_sha256||
   pg_catalog.uuid_send(baseline.organization_revision_id)||
   organizations.catalog_sha256||organizations.presentation_sha256||
   pg_catalog.uuid_send(baseline.role_revision_id)||
   roles.catalog_sha256||roles.presentation_sha256)
);
$function$;
CREATE FUNCTION system_resource_layout_repair_installation_defaults_v2(
    p_successor_contract text)
RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    current_record system_resource_layout_publications%ROWTYPE;
    v_revision_id uuid; v_created_at timestamptz; v_definition text;
    v_presentation text; v_definition_sha bytea; v_presentation_sha bytea;
    v_next_version bigint; v_changed bigint := 0;
BEGIN
 IF p_successor_contract IS DISTINCT FROM 'curated-installation-default-successor.v1' THEN
  RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='system resource successor contract is invalid';
 END IF;
 IF (SELECT pg_catalog.count(*) FROM system_resource_layout_installation_baselines)=0 THEN RETURN 0; END IF;
 IF NOT system_resource_layout_installation_baseline_evidence_valid_v2(
    'installation-baseline-evidence.v1') THEN
  RAISE EXCEPTION USING ERRCODE='23514', MESSAGE='system resource installation baseline evidence is invalid';
 END IF;
 FOR current_record IN SELECT publication.* FROM
  system_resource_layout_publications publication
  JOIN system_resource_layout_installation_baselines baseline
   ON publication.revision_id=CASE publication.resource_kind
    WHEN 'user' THEN baseline.user_revision_id
    WHEN 'organization' THEN baseline.organization_revision_id
    ELSE baseline.role_revision_id END
  WHERE publication.generation=1 ORDER BY publication.resource_kind
 LOOP
  SELECT greatest(current_record.updated_at,revision.created_at)+INTERVAL '1 microsecond'
   INTO v_created_at FROM system_resource_layout_revisions AS revision
   WHERE revision.id=current_record.revision_id;
  v_revision_id:=system_resource_layout_deterministic_uuid_v7_v1(
   current_record.revision_id::text,p_successor_contract,v_created_at);
  v_definition:=system_resource_layout_curated_definition_v1(
   current_record.resource_kind,current_record.catalog_json);
  v_definition_sha:=pg_catalog.sha256(pg_catalog.convert_to(v_definition,'UTF8'));
  v_presentation:=system_resource_layout_compile_curated_v1(
   v_revision_id,v_created_at,v_definition,current_record.catalog_json);
  v_presentation_sha:=pg_catalog.sha256(pg_catalog.convert_to(v_presentation,'UTF8'));
  SELECT COALESCE(pg_catalog.max(revision.version),0)+1 INTO v_next_version
    FROM system_resource_layout_revisions AS revision
   WHERE revision.resource_kind=current_record.resource_kind;
  INSERT INTO system_resource_layout_revisions
   (id,resource_kind,version,state,definition_json,definition_sha256,
    created_at,updated_at,published_at,row_revision)
  VALUES(v_revision_id,current_record.resource_kind,v_next_version,'draft',
   v_definition,v_definition_sha,v_created_at,v_created_at,NULL,1);
  UPDATE system_resource_layout_revisions AS revision SET state='published',
   published_at=v_created_at,updated_at=v_created_at,row_revision=2
   WHERE revision.id=v_revision_id AND revision.state='draft'
    AND revision.row_revision=1;
  IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='40001', MESSAGE='system resource curated successor publish conflicted'; END IF;
  UPDATE system_resource_layout_publications AS publication
   SET revision_id=v_revision_id,
   generation=current_record.generation+1,updated_at=v_created_at,
   source_definition_sha256=v_definition_sha,
   compiler_version='system-resource-presentation-compiler.v2',
   catalog_json=current_record.catalog_json,
   catalog_sha256=current_record.catalog_sha256,
   presentation_json=v_presentation,presentation_sha256=v_presentation_sha
   WHERE publication.resource_kind=current_record.resource_kind
    AND publication.revision_id=current_record.revision_id
    AND publication.generation=current_record.generation;
  IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='40001',
   MESSAGE='system resource curated publication conflicted'; END IF;
  v_changed:=v_changed+1;
 END LOOP;
 RETURN v_changed;
END;
$function$;
SELECT system_resource_layout_repair_installation_defaults_v2('curated-installation-default-successor.v1');
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_installation_catalog_valid_v2(text,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION system_resource_layout_installation_baseline_evidence_valid_v2(text) FROM PUBLIC;
REVOKE ALL ON FUNCTION system_resource_layout_repair_installation_defaults_v2(text) FROM PUBLIC;
