-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_apply_v1_campaign_v1(
    p_operation_id uuid, p_run_id uuid, p_record jsonb
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_discount jsonb;
    v_metadata jsonb;
    v_sha bytea;
BEGIN
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
      AND run.scope_kind = 'global' AND run.state = 'planned' AND run.eligible;
    v_metadata := pg_catalog.jsonb_build_object(
        'name', p_record->>'name', 'description', p_record->>'description',
        'banner_text', p_record->>'banner_text',
        'auto_apply', (p_record->>'auto_apply')::boolean
    );
    INSERT INTO platform.billing_campaigns (
        id, authority_id, campaign_key, lifecycle, created_at, retired_at, revision,
        legacy_created_by
    ) VALUES (
        (p_record->>'id')::uuid, v_run.authority_id, p_record->>'key',
        CASE WHEN p_record->>'retired_at' IS NULL THEN 'active' ELSE 'retired' END,
        (p_record->>'created_at')::timestamptz,
        NULLIF(p_record->>'retired_at', '')::timestamptz, 1,
        (p_record->>'created_by')::uuid
    );
    INSERT INTO platform.billing_campaign_versions (
        id, authority_id, campaign_id, version, priority,
        expected_discount_count, starts_at, ends_at,
        presentation_metadata, presentation_sha256, publication_state,
        publication_operation_id, published_at, presentation_name,
        presentation_description, presentation_banner_text, auto_apply, available
    ) VALUES (
        (p_record->>'version_id')::uuid, v_run.authority_id,
        (p_record->>'id')::uuid, 1, (p_record->>'priority')::integer,
        pg_catalog.jsonb_array_length(p_record->'discount_version_ids'),
        (p_record->>'starts_at')::timestamptz,
        NULLIF(p_record->>'ends_at', '')::timestamptz,
        v_metadata, pg_catalog.sha256(pg_catalog.convert_to(v_metadata::text, 'UTF8')),
        'published', (p_record->>'publication_operation_id')::uuid,
        (p_record->>'updated_at')::timestamptz, p_record->>'name',
        p_record->>'description', p_record->>'banner_text',
        (p_record->>'auto_apply')::boolean, (p_record->>'is_active')::boolean
    );
    FOR v_discount IN
        SELECT value
        FROM pg_catalog.jsonb_array_elements(p_record->'discount_version_ids')
    LOOP
        INSERT INTO platform.billing_campaign_discounts (
            authority_id, campaign_version_id, discount_version_id, created_at
        ) VALUES (
            v_run.authority_id, (p_record->>'version_id')::uuid,
            (v_discount #>> '{}')::uuid, (p_record->>'created_at')::timestamptz
        );
    END LOOP;
    v_sha := platform.billing_v1_adoption_catalog_fingerprint_v1(
        'campaign', (p_record->>'id')::uuid
    );
    PERFORM platform.billing_record_v1_adoption_mapping_v1(
        p_operation_id, p_run_id, 'campaigns', p_record->>'source_id',
        'campaign', (p_record->>'id')::uuid, v_sha, v_run.snapshot_at
    );
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_apply_v1_campaign_v1(
    uuid, uuid, jsonb
) FROM PUBLIC;
-- +goose StatementEnd
