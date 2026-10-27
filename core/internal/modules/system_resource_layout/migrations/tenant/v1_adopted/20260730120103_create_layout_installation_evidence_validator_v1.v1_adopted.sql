-- +goose Up
-- Migration stream version: 20260730120103.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_installation_catalog_valid_v1(
    p_resource_kind text, p_catalog_json text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_catalog jsonb;
    v_prefix text;
    v_static_count integer;
    v_ids jsonb;
    v_types jsonb;
    v_editable jsonb;
    v_unique bigint;
    v_invalid bigint;
BEGIN
    IF p_resource_kind NOT IN ('user','organization','role') THEN
        RETURN false;
    END IF;
    v_catalog := p_catalog_json::jsonb;
    v_prefix := CASE p_resource_kind WHEN 'user' THEN 'a1000000-'
        WHEN 'role' THEN 'b1000000-' ELSE 'c1000000-' END;
    v_static_count := CASE p_resource_kind
        WHEN 'user' THEN 11 WHEN 'role' THEN 14 ELSE 11 END;
    SELECT pg_catalog.jsonb_agg(
        v_prefix || '0000-4000-8000-' ||
        pg_catalog.lpad(pg_catalog.to_hex(series.number), 12, '0')
        ORDER BY series.number
    ) INTO v_ids
    FROM pg_catalog.generate_series(1, v_static_count) AS series(number);
    v_types := CASE p_resource_kind
        WHEN 'user' THEN
            '["uuid","text","text","enum","enum","uuid_list","uuid_list","uuid","integer","timestamp","timestamp"]'::jsonb
        WHEN 'role' THEN
            '["uuid","enum","uuid","relation","text","text","enum","boolean","boolean","uuid_list","uuid","integer","timestamp","timestamp"]'::jsonb
        ELSE
            '["uuid","text","text","enum","relation","uuid_list","uuid_list","uuid","integer","timestamp","timestamp"]'::jsonb
        END;
    v_editable := '["user:002","user:006","user:007","role:002","role:003","role:004","role:005","role:006","role:009","organization:002","organization:003","organization:005"]'::jsonb;
    IF v_catalog IS NULL
       OR v_catalog ->> 'contract' IS DISTINCT FROM
          'forms-presentation-field-catalog.v1'
       OR pg_catalog.jsonb_typeof(v_catalog -> 'fields') <> 'array'
       OR (SELECT pg_catalog.count(*)
           FROM pg_catalog.jsonb_object_keys(v_catalog)) <> 2
       OR pg_catalog.jsonb_array_length(v_catalog -> 'fields') <>
          v_static_count + 1 THEN
        RETURN false;
    END IF;
    SELECT pg_catalog.count(DISTINCT item.id),
           pg_catalog.count(*) FILTER (WHERE
               (SELECT pg_catalog.count(*)
                FROM pg_catalog.jsonb_object_keys(item.value)) <> 4
            OR (item.ordinality <= v_static_count
                AND item.id IS DISTINCT FROM
                    v_ids ->> (item.ordinality::integer - 1))
            OR (item.ordinality = v_static_count + 1
                AND (item.id LIKE v_prefix || '%'
                     OR item.id !~
                       '^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'))
            OR item.value ->> 'type_id' IS DISTINCT FROM
                CASE WHEN item.ordinality = v_static_count + 1 THEN 'text'
                ELSE v_types ->> (item.ordinality::integer - 1) END
            OR item.value ->> 'value_profile' IS DISTINCT FROM
                CASE item.value ->> 'type_id'
                WHEN 'text' THEN 'text_nfc_v1'
                WHEN 'uuid' THEN 'uuid_rfc4122_lower_v1'
                WHEN 'enum' THEN 'enum_code_v1'
                WHEN 'uuid_list' THEN 'uuid_list_v1'
                WHEN 'integer' THEN 'integer_base10_v1'
                WHEN 'timestamp' THEN 'timestamp_utc_microsecond_v1'
                WHEN 'relation' THEN 'relation_uuid_v1'
                WHEN 'boolean' THEN 'boolean_v1' END
            OR item.value ->> 'mode' IS DISTINCT FROM
                CASE WHEN item.ordinality = v_static_count + 1
                       OR v_editable ? (
                           p_resource_kind || ':' ||
                           pg_catalog.right(item.id, 3)
                       )
                THEN 'editable' ELSE 'read_only' END)
    INTO v_unique, v_invalid
    FROM (
        SELECT field.value, field.value ->> 'id' AS id, field.ordinality
        FROM pg_catalog.jsonb_array_elements(v_catalog -> 'fields')
            WITH ORDINALITY AS field(value, ordinality)
    ) AS item;
    RETURN v_unique = v_static_count + 1 AND v_invalid = 0;
END;
$function$;

CREATE FUNCTION system_resource_layout_installation_baseline_evidence_valid_v1(
    p_contract text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
RETURN p_contract IS NOT DISTINCT FROM
    'installation-baseline-evidence.v1'
AND NOT EXISTS (
    SELECT 1
    FROM system_resource_layout_installation_baselines AS baseline
    CROSS JOIN (VALUES ('user'),('organization'),('role')) AS expected(kind)
    LEFT JOIN system_resource_layout_publications AS publication
      ON publication.resource_kind = expected.kind
    LEFT JOIN system_resource_layout_revisions AS revision
      ON revision.id = CASE expected.kind
         WHEN 'user' THEN baseline.user_revision_id
         WHEN 'organization' THEN baseline.organization_revision_id
         ELSE baseline.role_revision_id END
    WHERE publication.resource_kind IS NULL
       OR publication.generation = 1 AND (
          publication.revision_id IS DISTINCT FROM revision.id
       OR revision.resource_kind IS DISTINCT FROM expected.kind
       OR revision.state IS DISTINCT FROM 'published'
       OR publication.catalog_contract IS DISTINCT FROM
          'forms-presentation-field-catalog.v1'
       OR publication.presentation_contract IS DISTINCT FROM
          'forms-presentation.v1'
       OR publication.compiler_version IS DISTINCT FROM
          'system-resource-presentation-compiler.v2'
       OR revision.definition_json::jsonb ->> 'contract' IS DISTINCT FROM
          'forgecrud/system-resource-layout/v2'
       OR publication.catalog_sha256 IS DISTINCT FROM pg_catalog.sha256(
          pg_catalog.convert_to(publication.catalog_json, 'UTF8'))
       OR publication.presentation_sha256 IS DISTINCT FROM pg_catalog.sha256(
          pg_catalog.convert_to(publication.presentation_json, 'UTF8'))
       OR revision.definition_sha256 IS DISTINCT FROM pg_catalog.sha256(
          pg_catalog.convert_to(revision.definition_json, 'UTF8'))
       OR publication.source_definition_sha256 IS DISTINCT FROM
          revision.definition_sha256
       OR NOT system_resource_layout_installation_catalog_valid_v1(
          expected.kind, publication.catalog_json))
) AND NOT EXISTS (
    SELECT 1
    FROM system_resource_layout_installation_baselines AS baseline
    JOIN system_resource_layout_publications AS users
      ON users.resource_kind = 'user'
    JOIN system_resource_layout_publications AS organizations
      ON organizations.resource_kind = 'organization'
    JOIN system_resource_layout_publications AS roles
      ON roles.resource_kind = 'role'
    WHERE users.generation = 1
      AND organizations.generation = 1
      AND roles.generation = 1
      AND baseline.baseline_sha256 IS DISTINCT FROM pg_catalog.sha256(
          pg_catalog.uuid_send(baseline.operation_id) ||
          pg_catalog.uuid_send(baseline.tenant_id) ||
          pg_catalog.uuid_send(baseline.user_revision_id) ||
          users.catalog_sha256 || users.presentation_sha256 ||
          pg_catalog.uuid_send(baseline.organization_revision_id) ||
          organizations.catalog_sha256 || organizations.presentation_sha256 ||
          pg_catalog.uuid_send(baseline.role_revision_id) ||
          roles.catalog_sha256 || roles.presentation_sha256)
);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    system_resource_layout_installation_catalog_valid_v1(text,text)
    FROM PUBLIC;
REVOKE ALL ON FUNCTION
    system_resource_layout_installation_baseline_evidence_valid_v1(text)
    FROM PUBLIC;
