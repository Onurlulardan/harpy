-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_campaign_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        IF NEW.lifecycle IS DISTINCT FROM 'active'
           OR NEW.revision IS DISTINCT FROM 1
           OR NEW.retired_at IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'billing campaign initial state is invalid';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing campaign cannot be deleted';
    END IF;
    IF OLD.lifecycle IS DISTINCT FROM 'active'
       OR NEW.id IS DISTINCT FROM OLD.id
       OR NEW.authority_id IS DISTINCT FROM OLD.authority_id
       OR NEW.campaign_key IS DISTINCT FROM OLD.campaign_key
       OR NEW.created_at IS DISTINCT FROM OLD.created_at
       OR NEW.revision IS DISTINCT FROM OLD.revision + 1
       OR NOT (
           (NEW.lifecycle = 'active' AND NEW.retired_at IS NULL)
           OR (NEW.lifecycle = 'retired' AND NEW.retired_at >= OLD.created_at)
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing campaign transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_campaign_v1() FROM PUBLIC;

CREATE FUNCTION platform.billing_select_active_campaign_v1(
    p_authority_id uuid,
    p_instant timestamptz
)
RETURNS TABLE (
    campaign_id uuid,
    campaign_key text,
    campaign_version_id uuid,
    campaign_version bigint,
    campaign_priority integer,
    starts_at timestamptz,
    ends_at timestamptz,
    presentation_metadata jsonb,
    presentation_sha256 bytea,
    discount_version_id uuid,
    discount_priority integer
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF p_instant IS NULL OR NOT pg_catalog.isfinite(p_instant) THEN
        RAISE EXCEPTION USING ERRCODE = '22007',
            MESSAGE = 'billing campaign selection instant is invalid';
    END IF;
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    RETURN QUERY
    WITH selected AS MATERIALIZED (
        SELECT root.id, root.campaign_key, version_row.id AS version_id,
               version_row.version, version_row.priority, version_row.starts_at,
               version_row.ends_at, version_row.presentation_metadata,
               version_row.presentation_sha256
        FROM platform.billing_campaigns AS root
        JOIN platform.billing_campaign_versions AS version_row
          ON version_row.authority_id = root.authority_id
         AND version_row.campaign_id = root.id
        WHERE root.authority_id = p_authority_id
          AND root.lifecycle = 'active'
          AND version_row.publication_state = 'published'
          AND version_row.starts_at <= p_instant
          AND (version_row.ends_at IS NULL OR p_instant < version_row.ends_at)
        ORDER BY version_row.priority DESC, version_row.id ASC
        LIMIT 1
    )
    SELECT selected.id, selected.campaign_key, selected.version_id,
           selected.version, selected.priority, selected.starts_at,
           selected.ends_at, selected.presentation_metadata,
           selected.presentation_sha256, eligible.id, eligible.priority
    FROM selected
    LEFT JOIN LATERAL (
        SELECT discount_version.id, discount_version.priority
        FROM platform.billing_campaign_discounts AS binding
        JOIN platform.billing_discount_versions AS discount_version
          ON discount_version.authority_id = binding.authority_id
         AND discount_version.id = binding.discount_version_id
        JOIN platform.billing_discounts AS discount_root
          ON discount_root.authority_id = discount_version.authority_id
         AND discount_root.id = discount_version.discount_id
        WHERE binding.authority_id = p_authority_id
          AND binding.campaign_version_id = selected.version_id
          AND discount_root.lifecycle = 'active'
          AND discount_version.valid_start <= p_instant
          AND (discount_version.valid_end IS NULL
               OR p_instant < discount_version.valid_end)
        ORDER BY discount_version.priority ASC, discount_version.id ASC
    ) AS eligible ON true
    ORDER BY eligible.priority ASC, eligible.id ASC NULLS LAST;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_select_active_campaign_v1(
    uuid, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_select_active_campaign_v1(
    uuid, timestamptz
) TO harpy_runtime;

CREATE TRIGGER billing_campaigns_transition_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_campaigns
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_campaign_v1();
ALTER TABLE platform.billing_campaigns
    ENABLE ALWAYS TRIGGER billing_campaigns_transition_guard;
-- +goose StatementEnd
