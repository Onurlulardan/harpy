-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_purge_invoice_retention_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_now timestamptz,
    p_limit integer
)
RETURNS TABLE (records_purged bigint, has_remaining boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    candidate_ids text;
    removed_count bigint;
    statement_count bigint;
    cutoff timestamptz;
BEGIN
    removed_count := 0;
    statement_count := 0;
    cutoff := p_now - INTERVAL '10 years';
    IF NOT platform.billing_retention_delete_allowed_v1(
        p_authority_id, p_tenant_id
    ) OR p_limit NOT BETWEEN 1 AND 1000 OR NOT pg_catalog.isfinite(p_now) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing invoice retention context is invalid';
    END IF;
    SELECT pg_catalog.string_agg(candidate.id::text, ',') INTO candidate_ids
    FROM (
        SELECT invoice.id
        FROM platform.billing_invoices AS invoice
        WHERE invoice.tenant_id = p_tenant_id AND invoice.state = 'finalized'
          AND invoice.finalized_at <= cutoff
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_settlements AS settlement
              WHERE settlement.tenant_id = p_tenant_id
                AND settlement.invoice_id = invoice.id
                AND settlement.occurred_at > cutoff
          )
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_adjustments AS adjustment
              WHERE adjustment.tenant_id = p_tenant_id
                AND adjustment.invoice_id = invoice.id
                AND adjustment.created_at > cutoff
          )
          AND NOT EXISTS (
              SELECT 1
              FROM platform.billing_provider_objects AS object_row
              JOIN platform.billing_webhook_receipts AS receipt
                ON receipt.tenant_id = object_row.tenant_id
               AND receipt.provider_object_mapping_id = object_row.id
              WHERE object_row.tenant_id = p_tenant_id
                AND object_row.invoice_id = invoice.id
                AND (receipt.retain_until > p_now OR receipt.processor_state IN (
                    'received', 'processing'
                ))
          )
          AND NOT EXISTS (
              SELECT 1
              FROM platform.billing_provider_objects AS object_row
              JOIN platform.billing_provider_events AS event_row
                ON event_row.tenant_id = object_row.tenant_id
               AND event_row.provider_object_mapping_id = object_row.id
              WHERE object_row.tenant_id = p_tenant_id
                AND object_row.invoice_id = invoice.id
                AND event_row.occurred_at > cutoff
          )
          AND NOT EXISTS (
              SELECT 1
              FROM platform.billing_provider_objects AS object_row
              JOIN platform.billing_provider_events AS event_row
                ON event_row.tenant_id = object_row.tenant_id
               AND event_row.provider_object_mapping_id = object_row.id
              JOIN platform.billing_subscription_boundary_evidence AS evidence
                ON evidence.tenant_id = event_row.tenant_id
               AND evidence.provider_event_id = event_row.id
              WHERE object_row.tenant_id = p_tenant_id
                AND object_row.invoice_id = invoice.id
                AND evidence.recorded_at > cutoff
          )
          AND NOT EXISTS (
              SELECT 1
              FROM platform.billing_provider_objects AS object_row
              JOIN platform.billing_provider_commands AS command_row
                ON command_row.tenant_id = object_row.tenant_id
               AND command_row.provider_object_mapping_id = object_row.id
              WHERE object_row.tenant_id = p_tenant_id
                AND object_row.invoice_id = invoice.id
                AND (command_row.completed_at IS NULL
                    OR command_row.completed_at > cutoff)
          )
        ORDER BY invoice.finalized_at, invoice.id
        FOR UPDATE SKIP LOCKED LIMIT p_limit
    ) AS candidate;
    IF candidate_ids IS NULL THEN
        RETURN QUERY SELECT 0::bigint, false;
        RETURN;
    END IF;

    DELETE FROM platform.billing_subscription_boundary_evidence AS evidence
    USING platform.billing_provider_events AS event_row,
          platform.billing_provider_objects AS object_row
    WHERE evidence.tenant_id = p_tenant_id
      AND evidence.provider_event_id = event_row.id
      AND event_row.tenant_id = p_tenant_id
      AND event_row.provider_object_mapping_id = object_row.id
      AND object_row.tenant_id = p_tenant_id
      AND object_row.invoice_id = ANY(
          pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
      );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;

    DELETE FROM platform.billing_provider_events AS event_row
    USING platform.billing_provider_objects AS object_row
    WHERE event_row.tenant_id = p_tenant_id
      AND event_row.provider_object_mapping_id = object_row.id
      AND object_row.tenant_id = p_tenant_id
      AND object_row.invoice_id = ANY(
          pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
      );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;

    DELETE FROM platform.billing_webhook_receipts AS receipt
    USING platform.billing_provider_objects AS object_row
    WHERE receipt.tenant_id = p_tenant_id
      AND receipt.provider_object_mapping_id = object_row.id
      AND object_row.tenant_id = p_tenant_id
      AND object_row.invoice_id = ANY(
          pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
      );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;

    DELETE FROM platform.billing_provider_commands AS command_row
    USING platform.billing_provider_objects AS object_row
    WHERE command_row.tenant_id = p_tenant_id
      AND command_row.provider_object_mapping_id = object_row.id
      AND object_row.tenant_id = p_tenant_id
      AND object_row.invoice_id = ANY(
          pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
      )
      AND command_row.state IN ('succeeded', 'failed', 'reconcile_required')
      AND command_row.completed_at <= cutoff;
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;

    DELETE FROM platform.billing_provider_objects AS object_row
    WHERE object_row.tenant_id = p_tenant_id
      AND object_row.invoice_id = ANY(
          pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
      );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;

    DELETE FROM platform.billing_adjustments
    WHERE tenant_id = p_tenant_id AND invoice_id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;
    DELETE FROM platform.billing_invoice_finalizations
    WHERE tenant_id = p_tenant_id AND invoice_id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;
    DELETE FROM platform.billing_settlements
    WHERE tenant_id = p_tenant_id AND invoice_id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;
    DELETE FROM platform.billing_invoice_lines
    WHERE tenant_id = p_tenant_id AND invoice_id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;
    DELETE FROM platform.billing_invoices
    WHERE tenant_id = p_tenant_id AND id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;

    RETURN QUERY SELECT removed_count,
        pg_catalog.cardinality(
            pg_catalog.string_to_array(candidate_ids, ',')
        ) = p_limit;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_purge_invoice_retention_v1(
    uuid,uuid,timestamptz,integer
) FROM PUBLIC;
-- +goose StatementEnd
