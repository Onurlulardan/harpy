-- +goose Up
-- Migration stream version: 20260730120102.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_compile_curated_v1(
    p_revision_id uuid, p_created_at timestamptz,
    p_definition_json text, p_catalog_json text
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_definition jsonb;
    v_catalog jsonb;
    v_surface jsonb;
    v_section jsonb;
    v_field jsonb;
    v_kind text;
    v_field_id text;
    v_mode text;
    v_interaction text;
    v_widget text;
    v_fields jsonb;
    v_surfaces jsonb;
    v_surface_translations jsonb;
    v_section_translations jsonb;
    v_seed text;
BEGIN
    v_definition := p_definition_json::jsonb;
    v_catalog := p_catalog_json::jsonb;
    v_surfaces := '[]'::jsonb;
    v_seed := p_revision_id::text;
    IF p_revision_id IS NULL OR p_created_at IS NULL
       OR v_definition ->> 'contract' <> 'forgecrud/system-resource-layout/v2'
       OR v_catalog ->> 'contract' <> 'forms-presentation-field-catalog.v1' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'system resource curated compiler input is invalid';
    END IF;
    FOR v_surface IN
        SELECT value FROM pg_catalog.jsonb_array_elements(
            v_definition -> 'surfaces'
        )
    LOOP
        v_kind := v_surface ->> 'kind';
        v_section := v_surface -> 'sections' -> 0;
        v_fields := '[]'::jsonb;
        SELECT COALESCE(pg_catalog.jsonb_agg(
            pg_catalog.jsonb_build_object(
                'label', translation.value ->> 'label',
                'locale', translation.value ->> 'locale'
            ) ORDER BY translation.value ->> 'locale'
        ), '[]'::jsonb)
        INTO v_surface_translations
        FROM pg_catalog.jsonb_array_elements(
            v_surface -> 'translations'
        ) AS translation(value);
        SELECT COALESCE(pg_catalog.jsonb_agg(
            pg_catalog.jsonb_build_object(
                'label', translation.value ->> 'label',
                'locale', translation.value ->> 'locale'
            ) ORDER BY translation.value ->> 'locale'
        ), '[]'::jsonb)
        INTO v_section_translations
        FROM pg_catalog.jsonb_array_elements(
            v_section -> 'translations'
        ) AS translation(value);
        FOR v_field IN
            SELECT value FROM pg_catalog.jsonb_array_elements(
                v_section -> 'fields'
            ) ORDER BY (value ->> 'order')::bigint, value ->> 'field_id'
        LOOP
            v_field_id := v_field ->> 'field_id';
            SELECT field.value ->> 'mode' INTO v_mode
            FROM pg_catalog.jsonb_array_elements(
                v_catalog -> 'fields'
            ) AS field(value)
            WHERE field.value ->> 'id' = v_field_id;
            IF v_mode IS NULL THEN
                RAISE EXCEPTION USING ERRCODE = '23514',
                    MESSAGE = 'system resource curated field is outside catalog';
            END IF;
            v_interaction := v_field ->> 'interaction';
            IF (v_kind IN ('record_detail','record_list','record_filter')
                OR v_mode = 'read_only') AND v_interaction <> 'disabled' THEN
                v_interaction := 'read_only';
            END IF;
            v_widget := CASE
                WHEN v_kind IN ('record_detail','record_list','record_filter')
                    THEN 'auto'
                ELSE v_field ->> 'widget'
            END;
            v_fields := v_fields || pg_catalog.jsonb_build_array(
                pg_catalog.jsonb_build_object(
                    'field_id',v_field_id,
                    'id',system_resource_layout_deterministic_uuid_v7_v1(
                        v_seed,v_kind || ':field:' || v_field_id,p_created_at
                    )::text,
                    'interaction',v_interaction,'kind','field','span',1,
                    'translations',(
                        SELECT COALESCE(pg_catalog.jsonb_agg(
                            pg_catalog.jsonb_build_object(
                                'label', translation.value ->> 'label',
                                'locale', translation.value ->> 'locale'
                            ) ORDER BY translation.value ->> 'locale'
                        ), '[]'::jsonb)
                        FROM pg_catalog.jsonb_array_elements(
                            v_field -> 'translations'
                        ) AS translation(value)
                    ),
                    'widget',v_widget
                )
            );
        END LOOP;
        v_surfaces := v_surfaces || pg_catalog.jsonb_build_array(
            pg_catalog.jsonb_build_object(
                'id',system_resource_layout_deterministic_uuid_v7_v1(
                    v_seed,v_kind || ':surface',p_created_at
                )::text,
                'kind',v_kind,
                'root',pg_catalog.jsonb_build_object(
                    'children',pg_catalog.jsonb_build_array(
                        pg_catalog.jsonb_build_object(
                            'children',v_fields,
                            'id',system_resource_layout_deterministic_uuid_v7_v1(
                                v_seed,v_kind || ':section:' ||
                                (v_section ->> 'id'),p_created_at
                            )::text,
                            'kind','container','span',1,
                            'translations',v_section_translations,
                            'variant','section'
                        )
                    ),
                    'id',system_resource_layout_deterministic_uuid_v7_v1(
                        v_seed,v_kind || ':root',p_created_at
                    )::text,
                    'kind','container','span',1,
                    'translations','[]'::jsonb,'variant','stack'
                ),
                'translations',v_surface_translations
            )
        );
    END LOOP;
    RETURN system_resource_layout_canonical_json_v1(
        pg_catalog.jsonb_build_object(
            'default_locale','en','locales',pg_catalog.jsonb_build_array('en','tr'),
            'registry_contract','forms-presentation-registry.v1',
            'surfaces',v_surfaces
        )
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_compile_curated_v1(
    uuid,timestamptz,text,text
) FROM PUBLIC;
