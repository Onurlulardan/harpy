-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_add_campaign_discount_v1(
    p_authority_id uuid, p_campaign_version_id uuid,
    p_discount_version_id uuid, p_created_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    INSERT INTO platform.billing_campaign_discounts (
        authority_id, campaign_version_id, discount_version_id, created_at
    ) VALUES (
        p_authority_id, p_campaign_version_id, p_discount_version_id, p_created_at
    );
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_add_campaign_discount_v1(
    uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_add_campaign_discount_v1(
    uuid,uuid,uuid,timestamptz
) TO forgecrud_runtime;
-- +goose StatementEnd
