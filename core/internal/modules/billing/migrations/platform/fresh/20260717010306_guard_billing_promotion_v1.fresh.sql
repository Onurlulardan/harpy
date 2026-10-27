-- +goose Up
-- Migration version: 20260717010306 UTC
-- +goose StatementBegin
CREATE FUNCTION platform.billing_assert_promotion_association_v1(
    p_authority_id uuid, p_operation_id uuid, p_discount_id uuid,
    p_old_discount_version_id uuid, p_new_discount_version_id uuid,
    p_campaign_id uuid, p_campaign_version_id_one uuid,
    p_campaign_version_id_two uuid, p_campaign_operation_id_one uuid,
    p_campaign_operation_id_two uuid, p_occurred_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF NOT pg_catalog.isfinite(p_occurred_at)
       OR pg_catalog.uuid_extract_version(p_operation_id) <> 7
       OR pg_catalog.uuid_extract_version(p_discount_id) <> 7
       OR pg_catalog.uuid_extract_version(p_new_discount_version_id) <> 7
       OR (p_old_discount_version_id IS NOT NULL
           AND pg_catalog.uuid_extract_version(p_old_discount_version_id) <> 7)
       OR (p_campaign_id IS NOT NULL
           AND pg_catalog.uuid_extract_version(p_campaign_id) <> 7)
       OR EXISTS (
           SELECT 1
           FROM pg_catalog.unnest(ARRAY[
               p_campaign_version_id_one, p_campaign_version_id_two,
               p_campaign_operation_id_one, p_campaign_operation_id_two
           ]) AS selected(identifier)
           WHERE pg_catalog.uuid_extract_version(selected.identifier) <> 7
       ) OR (
           SELECT pg_catalog.count(DISTINCT selected.identifier)
           FROM pg_catalog.unnest(ARRAY[
               p_campaign_version_id_one, p_campaign_version_id_two,
               p_campaign_operation_id_one, p_campaign_operation_id_two,
               p_operation_id, p_discount_id, p_new_discount_version_id
           ]) AS selected(identifier)
       ) <> 7 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing promotion association identity is invalid';
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM platform.billing_catalog_operations AS operation
        JOIN platform.billing_discount_versions AS version_row
          ON version_row.authority_id = operation.authority_id
         AND version_row.id = operation.result_version_id
         AND version_row.discount_id = operation.aggregate_id
        WHERE operation.operation_id = p_operation_id
          AND operation.authority_id = p_authority_id
          AND operation.aggregate_kind = 'discount'
          AND operation.aggregate_id = p_discount_id
          AND operation.result_version_id = p_new_discount_version_id
    ) OR (
        p_old_discount_version_id IS NOT NULL AND NOT EXISTS (
            SELECT 1
            FROM platform.billing_discount_versions AS version_row
            WHERE version_row.authority_id = p_authority_id
              AND version_row.id = p_old_discount_version_id
              AND version_row.discount_id = p_discount_id
        )
    ) OR (
        p_campaign_id IS NOT NULL AND NOT EXISTS (
            SELECT 1
            FROM platform.billing_campaigns AS campaign
            WHERE campaign.authority_id = p_authority_id
              AND campaign.id = p_campaign_id
              AND campaign.lifecycle = 'active'
        )
    ) OR EXISTS (
        SELECT 1
        FROM pg_catalog.unnest(ARRAY[
            p_campaign_version_id_one, p_campaign_version_id_two
        ]) AS selected(identifier)
        JOIN platform.billing_campaign_versions AS existing
          ON existing.id = selected.identifier
    ) OR EXISTS (
        SELECT 1
        FROM pg_catalog.unnest(ARRAY[
            p_campaign_operation_id_one, p_campaign_operation_id_two
        ]) AS selected(identifier)
        JOIN platform.billing_catalog_operations AS existing
          ON existing.operation_id = selected.identifier
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'billing promotion association authority is invalid';
    END IF;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_assert_promotion_association_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;

DROP FUNCTION platform.billing_select_active_campaign_v1(uuid, timestamptz);
CREATE FUNCTION platform.billing_select_active_campaign_v1(
    p_authority_id uuid, p_instant timestamptz
)
RETURNS TABLE (
    campaign_id uuid, campaign_key text, campaign_version_id uuid,
    campaign_version bigint, campaign_priority integer,
    starts_at timestamptz, ends_at timestamptz,
    presentation_metadata jsonb, presentation_sha256 bytea,
    discount_version_id uuid, discount_priority integer
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
          AND version_row.available
          AND version_row.starts_at <= p_instant
          AND (version_row.ends_at IS NULL OR p_instant < version_row.ends_at)
          AND NOT EXISTS (
              SELECT 1
              FROM platform.billing_campaign_versions AS successor
              WHERE successor.authority_id = version_row.authority_id
                AND successor.superseded_version_id = version_row.id
          )
        ORDER BY root.created_at DESC, root.id ASC
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
          AND discount_version.available
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
) TO forgecrud_runtime;

CREATE TRIGGER billing_promotion_association_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_promotion_association_receipts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_promotion_association_receipts
    ENABLE ALWAYS TRIGGER billing_promotion_association_authority_guard;

CREATE TRIGGER billing_promotion_association_immutable_guard
BEFORE UPDATE OR DELETE ON platform.billing_promotion_association_receipts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_promotion_association_receipts
    ENABLE ALWAYS TRIGGER billing_promotion_association_immutable_guard;
-- +goose StatementEnd
