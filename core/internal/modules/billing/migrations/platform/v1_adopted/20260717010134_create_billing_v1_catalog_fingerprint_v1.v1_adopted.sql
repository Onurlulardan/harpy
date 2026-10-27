-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_v1_adoption_catalog_fingerprint_v1(
    p_record_kind text,
    p_target_id uuid
)
RETURNS bytea
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_document jsonb;
BEGIN
    IF p_record_kind IS NOT DISTINCT FROM 'offering' THEN
        SELECT pg_catalog.jsonb_build_object(
            'root', pg_catalog.to_jsonb(root_row),
            'version', pg_catalog.to_jsonb(version_row),
            'prices', COALESCE((
                SELECT pg_catalog.jsonb_agg(
                    pg_catalog.to_jsonb(price_row) ORDER BY price_row.id
                )
                FROM platform.billing_price_versions AS price_row
                WHERE price_row.offering_version_id = version_row.id
            ), '[]'::jsonb),
            'grants', COALESCE((
                SELECT pg_catalog.jsonb_agg(
                    pg_catalog.to_jsonb(grant_row)
                    ORDER BY grant_row.entitlement_definition_id
                )
                FROM platform.billing_offering_entitlements AS grant_row
                WHERE grant_row.offering_version_id = version_row.id
            ), '[]'::jsonb)
        ) INTO v_document
        FROM platform.billing_offerings AS root_row
        JOIN platform.billing_offering_versions AS version_row
          ON version_row.authority_id = root_row.authority_id
         AND version_row.offering_id = root_row.id
         AND version_row.version = 1
        WHERE root_row.id = p_target_id;
    ELSIF p_record_kind IS NOT DISTINCT FROM 'discount' THEN
        SELECT pg_catalog.jsonb_build_object(
            'root', pg_catalog.to_jsonb(root_row),
            'version', pg_catalog.to_jsonb(version_row),
            'offerings', COALESCE((
                SELECT pg_catalog.jsonb_agg(
                    pg_catalog.to_jsonb(binding)
                    ORDER BY binding.offering_version_id
                )
                FROM platform.billing_discount_offerings AS binding
                WHERE binding.discount_version_id = version_row.id
            ), '[]'::jsonb)
        ) INTO v_document
        FROM platform.billing_discounts AS root_row
        JOIN platform.billing_discount_versions AS version_row
          ON version_row.authority_id = root_row.authority_id
         AND version_row.discount_id = root_row.id
         AND version_row.version = 1
        WHERE root_row.id = p_target_id;
    ELSIF p_record_kind IS NOT DISTINCT FROM 'campaign' THEN
        SELECT pg_catalog.jsonb_build_object(
            'root', pg_catalog.to_jsonb(root_row),
            'version', pg_catalog.to_jsonb(version_row),
            'discounts', COALESCE((
                SELECT pg_catalog.jsonb_agg(
                    pg_catalog.to_jsonb(binding)
                    ORDER BY binding.discount_version_id
                )
                FROM platform.billing_campaign_discounts AS binding
                WHERE binding.campaign_version_id = version_row.id
            ), '[]'::jsonb)
        ) INTO v_document
        FROM platform.billing_campaigns AS root_row
        JOIN platform.billing_campaign_versions AS version_row
          ON version_row.authority_id = root_row.authority_id
         AND version_row.campaign_id = root_row.id
         AND version_row.version = 1
        WHERE root_row.id = p_target_id;
    ELSIF p_record_kind IS NOT DISTINCT FROM 'provider_account' THEN
        SELECT pg_catalog.to_jsonb(account)
        INTO v_document
        FROM platform.billing_provider_accounts AS account
        WHERE account.id = p_target_id;
    ELSE
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing V1 catalog fingerprint kind is invalid';
    END IF;
    IF v_document IS NULL THEN
        RETURN NULL;
    END IF;
    RETURN pg_catalog.sha256(
        pg_catalog.convert_to(v_document::text, 'UTF8')
    );
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_v1_adoption_catalog_fingerprint_v1(
    text, uuid
) FROM PUBLIC;
-- +goose StatementEnd

