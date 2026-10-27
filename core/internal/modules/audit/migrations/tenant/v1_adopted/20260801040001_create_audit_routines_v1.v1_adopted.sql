-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION audit_entry_append_v1(
    p_id uuid, p_owner_audit_id uuid, p_actor_id uuid,
    p_organization_id uuid, p_correlation_id uuid,
    p_owner_module text, p_operation_id text, p_entity_type text,
    p_entity_id text, p_form_slug text, p_action text,
    p_authorization_version bigint, p_reason_code text,
    p_field_changes jsonb, p_event_sha256 bytea,
    p_occurred_at timestamptz, p_recorded_at timestamptz
) RETURNS TABLE (
    id uuid, owner_audit_id uuid, actor_id uuid, organization_id uuid,
    correlation_id uuid, owner_module text, operation_id text,
    entity_type text, entity_id text, form_slug text, action text,
    authorization_version bigint, reason_code text, field_changes jsonb,
    event_sha256 bytea, occurred_at timestamptz, recorded_at timestamptz,
    replay boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE
    v_entry audit_entries%ROWTYPE;
    v_rows bigint;
BEGIN
    INSERT INTO audit_entries (
        id, owner_audit_id, actor_id, organization_id, correlation_id,
        owner_module, operation_id, entity_type, entity_id, form_slug,
        action, authorization_version, reason_code, field_changes,
        event_sha256, occurred_at, recorded_at
    ) VALUES (
        p_id, p_owner_audit_id, p_actor_id, p_organization_id, p_correlation_id,
        p_owner_module, p_operation_id, p_entity_type, p_entity_id, p_form_slug,
        p_action, p_authorization_version, p_reason_code, p_field_changes,
        p_event_sha256, p_occurred_at, p_recorded_at
    ) ON CONFLICT ON CONSTRAINT audit_entries_owner_audit_id_key DO NOTHING;
    GET DIAGNOSTICS v_rows = ROW_COUNT;

    SELECT entries.* INTO STRICT v_entry
      FROM audit_entries AS entries
     WHERE entries.owner_audit_id = p_owner_audit_id;
    IF v_entry.event_sha256 IS DISTINCT FROM p_event_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6802', MESSAGE = 'AUDIT_OWNER_KEY_REUSED';
    END IF;

    RETURN QUERY SELECT
        v_entry.id, v_entry.owner_audit_id, v_entry.actor_id,
        v_entry.organization_id, v_entry.correlation_id, v_entry.owner_module,
        v_entry.operation_id, v_entry.entity_type, v_entry.entity_id,
        v_entry.form_slug, v_entry.action, v_entry.authorization_version,
        v_entry.reason_code, v_entry.field_changes, v_entry.event_sha256,
        v_entry.occurred_at, v_entry.recorded_at, v_rows = 0;
END;
$$;
CREATE FUNCTION audit_entries_list_v1(
    p_owner_module text, p_entity_type text, p_entity_id text,
    p_form_slug text, p_action text, p_actor_id uuid,
    p_organization_id uuid, p_correlation_id uuid,
    p_start_at timestamptz, p_end_at timestamptz, p_search text,
    p_page integer, p_limit integer
) RETURNS TABLE (
    id uuid, owner_audit_id uuid, actor_id uuid, organization_id uuid,
    correlation_id uuid, owner_module text, operation_id text,
    entity_type text, entity_id text, form_slug text, action text,
    authorization_version bigint, reason_code text, field_changes jsonb,
    event_sha256 bytea, occurred_at timestamptz, recorded_at timestamptz,
    total_count bigint
)
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path FROM CURRENT AS $$
BEGIN
    IF p_page NOT BETWEEN 1 AND 1000 OR p_limit NOT BETWEEN 1 AND 100 THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'AUDIT_PAGE_INVALID';
    END IF;
    RETURN QUERY
    SELECT entries.id, entries.owner_audit_id, entries.actor_id,
           entries.organization_id, entries.correlation_id, entries.owner_module,
           entries.operation_id, entries.entity_type, entries.entity_id,
           entries.form_slug, entries.action, entries.authorization_version,
           entries.reason_code, entries.field_changes, entries.event_sha256,
           entries.occurred_at, entries.recorded_at,
           pg_catalog.count(*) OVER ()::bigint
      FROM audit_entries AS entries
     WHERE (p_owner_module IS NULL OR entries.owner_module = p_owner_module)
       AND (p_entity_type IS NULL OR entries.entity_type = p_entity_type)
       AND (p_entity_id IS NULL OR entries.entity_id = p_entity_id)
       AND (p_form_slug IS NULL OR entries.form_slug = p_form_slug)
       AND (p_action IS NULL OR entries.action = p_action)
       AND (p_actor_id IS NULL OR entries.actor_id = p_actor_id)
       AND (p_organization_id IS NULL OR entries.organization_id = p_organization_id)
       AND (p_correlation_id IS NULL OR entries.correlation_id = p_correlation_id)
       AND (p_start_at IS NULL OR entries.occurred_at >= p_start_at)
       AND (p_end_at IS NULL OR entries.occurred_at <= p_end_at)
       AND (p_search IS NULL OR pg_catalog.strpos(
            pg_catalog.lower(
                entries.entity_id || ' ' || coalesce(entries.form_slug, '') ||
                ' ' || entries.owner_module
            ), pg_catalog.lower(p_search)) > 0)
     ORDER BY entries.occurred_at DESC, entries.id DESC
     OFFSET ((p_page - 1) * p_limit) LIMIT p_limit;
END;
$$;
CREATE FUNCTION audit_entry_get_v1(p_id uuid) RETURNS TABLE (
    id uuid, owner_audit_id uuid, actor_id uuid, organization_id uuid,
    correlation_id uuid, owner_module text, operation_id text,
    entity_type text, entity_id text, form_slug text, action text,
    authorization_version bigint, reason_code text, field_changes jsonb,
    event_sha256 bytea, occurred_at timestamptz, recorded_at timestamptz
)
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path FROM CURRENT AS $$
BEGIN
    RETURN QUERY
    SELECT entries.id, entries.owner_audit_id, entries.actor_id,
           entries.organization_id, entries.correlation_id, entries.owner_module,
           entries.operation_id, entries.entity_type, entries.entity_id,
           entries.form_slug, entries.action, entries.authorization_version,
           entries.reason_code, entries.field_changes, entries.event_sha256,
           entries.occurred_at, entries.recorded_at
      FROM audit_entries AS entries WHERE entries.id = p_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6801', MESSAGE = 'AUDIT_ENTRY_NOT_FOUND';
    END IF;
END;
$$;
CREATE FUNCTION audit_stats_v1(p_now timestamptz) RETURNS jsonb
LANGUAGE sql STABLE SECURITY DEFINER SET search_path FROM CURRENT AS $$
WITH action_groups AS (
    SELECT entries.action AS key, pg_catalog.count(*)::bigint AS value
      FROM audit_entries AS entries GROUP BY entries.action
), entity_groups AS (
    SELECT entries.entity_type AS key, pg_catalog.count(*)::bigint AS value
      FROM audit_entries AS entries GROUP BY entries.entity_type
), actor_groups AS (
    SELECT entries.actor_id, pg_catalog.count(*)::bigint AS value
      FROM audit_entries AS entries GROUP BY entries.actor_id
     ORDER BY value DESC, entries.actor_id LIMIT 10
), active_entities AS (
    SELECT entries.entity_type, entries.entity_id,
           coalesce(entries.form_slug, '') AS form_slug,
           pg_catalog.count(*)::bigint AS value
      FROM audit_entries AS entries
     GROUP BY entries.entity_type, entries.entity_id, entries.form_slug
     ORDER BY value DESC, entries.entity_type, entries.entity_id LIMIT 10
)
SELECT pg_catalog.jsonb_build_object(
    'total', (SELECT pg_catalog.count(*)::bigint FROM audit_entries),
    'last_24_hours', (SELECT pg_catalog.count(*)::bigint FROM audit_entries
                       WHERE occurred_at >= p_now - interval '24 hours'),
    'last_hour', (SELECT pg_catalog.count(*)::bigint FROM audit_entries
                   WHERE occurred_at >= p_now - interval '1 hour'),
    'by_action', coalesce((
        SELECT pg_catalog.jsonb_object_agg(key, value) FROM action_groups
    ), '{}'::jsonb),
    'by_entity', coalesce((
        SELECT pg_catalog.jsonb_object_agg(key, value) FROM entity_groups
    ), '{}'::jsonb),
    'actors', coalesce((
        SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
            'actor_id', actor_id, 'count', value) ORDER BY value DESC, actor_id)
          FROM actor_groups
    ), '[]'::jsonb),
    'entities', coalesce((
        SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
            'entity_type', entity_type, 'entity_id', entity_id,
            'form_slug', form_slug, 'count', value)
            ORDER BY value DESC, entity_type, entity_id)
          FROM active_entities
    ), '[]'::jsonb)
);
$$;
-- +goose StatementEnd
