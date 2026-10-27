-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_v1_adoption_tenant_fingerprint_v1(
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
    IF p_record_kind IS NOT DISTINCT FROM 'base_assignment' THEN
        SELECT pg_catalog.to_jsonb(row_value) INTO v_document
        FROM platform.billing_tenant_base_offerings AS row_value
        WHERE row_value.id = p_target_id;
    ELSIF p_record_kind IS NOT DISTINCT FROM 'subscription' THEN
        SELECT pg_catalog.jsonb_build_object(
            'subscription', pg_catalog.to_jsonb(row_value),
            'legacy_wire', pg_catalog.to_jsonb(wire_row)
        ) INTO v_document
        FROM platform.billing_subscriptions AS row_value
        LEFT JOIN platform.billing_v1_subscription_wire_archive AS wire_row
          ON wire_row.subscription_id = row_value.id
        WHERE row_value.id = p_target_id;
    ELSIF p_record_kind IS NOT DISTINCT FROM 'redemption' THEN
        SELECT pg_catalog.to_jsonb(row_value) INTO v_document
        FROM platform.billing_discount_redemptions AS row_value
        WHERE row_value.id = p_target_id;
    ELSIF p_record_kind IN ('usage_opening', 'gauge_opening') THEN
        SELECT pg_catalog.jsonb_build_object(
            'period', pg_catalog.to_jsonb(period_row),
            'total', pg_catalog.to_jsonb(total_row),
            'entries', COALESCE((
                SELECT pg_catalog.jsonb_agg(
                    pg_catalog.to_jsonb(entry_row) ORDER BY entry_row.id
                )
                FROM platform.billing_usage_entries AS entry_row
                WHERE entry_row.meter_period_id = period_row.id
                  AND entry_row.entry_kind = 'opening'
            ), '[]'::jsonb),
            'evidence', pg_catalog.to_jsonb(evidence_row)
        ) INTO v_document
        FROM platform.billing_meter_periods AS period_row
        JOIN platform.billing_usage_totals AS total_row
          ON total_row.meter_period_id = period_row.id
        JOIN platform.billing_v1_adoption_opening_evidence AS evidence_row
          ON evidence_row.meter_period_id = period_row.id
        WHERE period_row.id = p_target_id
          AND evidence_row.evidence_kind = CASE p_record_kind
              WHEN 'usage_opening' THEN 'submission_usage'
              ELSE 'owner_gauge'
          END;
    ELSIF p_record_kind IS NOT DISTINCT FROM 'legacy_event' THEN
        SELECT pg_catalog.to_jsonb(row_value) INTO v_document
        FROM platform.billing_v1_event_archive AS row_value
        WHERE row_value.id = p_target_id;
    ELSE
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing V1 tenant fingerprint kind is invalid';
    END IF;
    IF v_document IS NULL THEN
        RETURN NULL;
    END IF;
    RETURN pg_catalog.sha256(
        pg_catalog.convert_to(v_document::text, 'UTF8')
    );
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_v1_adoption_tenant_fingerprint_v1(
    text, uuid
) FROM PUBLIC;
-- +goose StatementEnd
