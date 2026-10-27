-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_purge_subscription_retention_v1(
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
    redemption_count bigint;
    candidate_count integer;
    cutoff timestamptz;
BEGIN
    removed_count := 0;
    statement_count := 0;
    redemption_count := 0;
    candidate_count := 0;
    cutoff := p_now - INTERVAL '10 years';
    IF NOT platform.billing_retention_delete_allowed_v1(
        p_authority_id, p_tenant_id
    ) OR p_limit NOT BETWEEN 1 AND 1000 OR NOT pg_catalog.isfinite(p_now) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing subscription retention context is invalid';
    END IF;
    WITH candidates AS (
        SELECT redemption.id
        FROM platform.billing_discount_redemptions AS redemption
        WHERE redemption.tenant_id = p_tenant_id
          AND redemption.relationship_ended_at <= cutoff
        ORDER BY redemption.relationship_ended_at, redemption.id
        FOR UPDATE SKIP LOCKED LIMIT p_limit
    ), removed AS (
        DELETE FROM platform.billing_discount_redemptions AS redemption
        USING candidates
        WHERE redemption.tenant_id = p_tenant_id
          AND redemption.id = candidates.id
        RETURNING 1
    ) SELECT pg_catalog.count(*) INTO redemption_count FROM removed;
    removed_count := redemption_count;

    SELECT pg_catalog.string_agg(candidate.id::text, ',') INTO candidate_ids
    FROM (
        SELECT subscription.id
        FROM platform.billing_subscriptions AS subscription
        WHERE subscription.tenant_id = p_tenant_id
          AND subscription.state IN ('cancelled', 'expired')
          AND subscription.updated_at <= cutoff
          AND platform.billing_subscription_retention_eligible_v1(
              p_tenant_id, subscription.id, cutoff, p_now
          )
        ORDER BY subscription.updated_at, subscription.id
        FOR UPDATE SKIP LOCKED LIMIT p_limit
    ) AS candidate;
    candidate_count := COALESCE(pg_catalog.cardinality(
        pg_catalog.string_to_array(candidate_ids, ',')
    ), 0);
    IF candidate_ids IS NULL THEN
        RETURN QUERY SELECT removed_count, redemption_count = p_limit;
        RETURN;
    END IF;

    DELETE FROM platform.billing_subscription_boundary_evidence
    WHERE tenant_id = p_tenant_id AND subscription_id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;

    DELETE FROM platform.billing_provider_events AS event_row
    USING platform.billing_provider_objects AS object_row
    WHERE event_row.tenant_id = p_tenant_id
      AND event_row.provider_object_mapping_id = object_row.id
      AND object_row.tenant_id = p_tenant_id
      AND object_row.subscription_id = ANY(
          pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
      );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;

    DELETE FROM platform.billing_webhook_receipts AS receipt
    USING platform.billing_provider_objects AS object_row
    WHERE receipt.tenant_id = p_tenant_id
      AND receipt.provider_object_mapping_id = object_row.id
      AND object_row.tenant_id = p_tenant_id
      AND object_row.subscription_id = ANY(
          pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
      );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;

    DELETE FROM platform.billing_provider_commands AS command_row
    WHERE command_row.tenant_id = p_tenant_id
      AND (command_row.subscription_id = ANY(
              pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
          )
        OR command_row.provider_object_mapping_id IN (
            SELECT object_row.id FROM platform.billing_provider_objects AS object_row
            WHERE object_row.tenant_id = p_tenant_id
              AND object_row.subscription_id = ANY(
                  pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
              )
        ));
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;

    DELETE FROM platform.billing_provider_objects
    WHERE tenant_id = p_tenant_id AND subscription_id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;

    DELETE FROM platform.billing_subscription_audit
    WHERE tenant_id = p_tenant_id AND subscription_id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;
    DELETE FROM platform.billing_subscription_boundary_jobs
    WHERE tenant_id = p_tenant_id AND subscription_id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;
    DELETE FROM platform.billing_subscription_changes
    WHERE tenant_id = p_tenant_id AND subscription_id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;
    DELETE FROM platform.billing_subscription_overrides
    WHERE tenant_id = p_tenant_id AND subscription_id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;
    DELETE FROM platform.billing_trial_claims
    WHERE tenant_id = p_tenant_id AND subscription_id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;
    DELETE FROM platform.billing_subscription_operations
    WHERE tenant_id = p_tenant_id AND subscription_id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;
    DELETE FROM platform.billing_discount_redemptions
    WHERE tenant_id = p_tenant_id AND subscription_id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;
    DELETE FROM platform.billing_subscriptions
    WHERE tenant_id = p_tenant_id AND id = ANY(
        pg_catalog.string_to_array(candidate_ids, ',')::uuid[]
    );
    GET DIAGNOSTICS statement_count = ROW_COUNT;
    removed_count := removed_count + statement_count;

    RETURN QUERY SELECT removed_count,
        redemption_count = p_limit OR candidate_count = p_limit;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_purge_subscription_retention_v1(
    uuid,uuid,timestamptz,integer
) FROM PUBLIC;
-- +goose StatementEnd
