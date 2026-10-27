-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_subscription_retention_eligible_v1(
    p_tenant_id uuid,
    p_subscription_id uuid,
    p_cutoff timestamptz,
    p_now timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    RETURN NOT EXISTS (
        SELECT 1 FROM platform.billing_invoices AS invoice
        WHERE invoice.tenant_id = p_tenant_id
          AND invoice.subscription_id = p_subscription_id
    ) AND NOT EXISTS (
        SELECT 1 FROM platform.billing_entitlement_decisions AS decision
        WHERE decision.tenant_id = p_tenant_id
          AND decision.subscription_id = p_subscription_id
    ) AND NOT EXISTS (
        SELECT 1 FROM platform.billing_discount_redemptions AS redemption
        WHERE redemption.tenant_id = p_tenant_id
          AND redemption.subscription_id = p_subscription_id
          AND (redemption.relationship_ended_at IS NULL
              OR redemption.relationship_ended_at > p_cutoff)
    ) AND NOT EXISTS (
        SELECT 1 FROM platform.billing_subscription_changes AS change_row
        WHERE change_row.tenant_id = p_tenant_id
          AND change_row.subscription_id = p_subscription_id
          AND (change_row.resolved_at IS NULL OR change_row.resolved_at > p_cutoff)
    ) AND NOT EXISTS (
        SELECT 1 FROM platform.billing_subscription_overrides AS override_row
        WHERE override_row.tenant_id = p_tenant_id
          AND override_row.subscription_id = p_subscription_id
          AND (override_row.effective_end IS NULL
              OR override_row.effective_end > p_cutoff)
    ) AND NOT EXISTS (
        SELECT 1 FROM platform.billing_subscription_operations AS operation_row
        WHERE operation_row.tenant_id = p_tenant_id
          AND operation_row.subscription_id = p_subscription_id
          AND operation_row.completed_at > p_cutoff
    ) AND NOT EXISTS (
        SELECT 1 FROM platform.billing_subscription_audit AS audit_row
        WHERE audit_row.tenant_id = p_tenant_id
          AND audit_row.subscription_id = p_subscription_id
          AND audit_row.occurred_at > p_cutoff
    ) AND NOT EXISTS (
        SELECT 1 FROM platform.billing_subscription_boundary_jobs AS job_row
        WHERE job_row.tenant_id = p_tenant_id
          AND job_row.subscription_id = p_subscription_id
          AND job_row.run_at > p_cutoff
    ) AND NOT EXISTS (
        SELECT 1 FROM platform.billing_subscription_boundary_evidence AS evidence
        WHERE evidence.tenant_id = p_tenant_id
          AND evidence.subscription_id = p_subscription_id
          AND evidence.recorded_at > p_cutoff
    ) AND NOT EXISTS (
        SELECT 1 FROM platform.billing_trial_claims AS trial_claim
        WHERE trial_claim.tenant_id = p_tenant_id
          AND trial_claim.subscription_id = p_subscription_id
          AND trial_claim.claimed_at > p_cutoff
    ) AND NOT EXISTS (
        SELECT 1
        FROM platform.billing_provider_objects AS object_row
        JOIN platform.billing_provider_events AS event_row
          ON event_row.tenant_id = object_row.tenant_id
         AND event_row.provider_object_mapping_id = object_row.id
        WHERE object_row.tenant_id = p_tenant_id
          AND object_row.subscription_id = p_subscription_id
          AND event_row.occurred_at > p_cutoff
    ) AND NOT EXISTS (
        SELECT 1
        FROM platform.billing_provider_objects AS object_row
        JOIN platform.billing_webhook_receipts AS receipt
          ON receipt.tenant_id = object_row.tenant_id
         AND receipt.provider_object_mapping_id = object_row.id
        WHERE object_row.tenant_id = p_tenant_id
          AND object_row.subscription_id = p_subscription_id
          AND (receipt.retain_until > p_now
              OR receipt.processor_state IN ('received', 'processing'))
    ) AND NOT EXISTS (
        SELECT 1 FROM platform.billing_provider_commands AS command_row
        WHERE command_row.tenant_id = p_tenant_id
          AND command_row.subscription_id = p_subscription_id
          AND (command_row.completed_at IS NULL
              OR command_row.completed_at > p_cutoff)
    ) AND NOT EXISTS (
        SELECT 1
        FROM platform.billing_provider_objects AS object_row
        JOIN platform.billing_provider_commands AS command_row
          ON command_row.tenant_id = object_row.tenant_id
         AND command_row.provider_object_mapping_id = object_row.id
        WHERE object_row.tenant_id = p_tenant_id
          AND object_row.subscription_id = p_subscription_id
          AND (command_row.completed_at IS NULL
              OR command_row.completed_at > p_cutoff)
    );
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_subscription_retention_eligible_v1(
    uuid,uuid,timestamptz,timestamptz
) FROM PUBLIC;
-- +goose StatementEnd
