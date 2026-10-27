-- +goose Up
CREATE TRIGGER billing_adjustments_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_adjustments
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_adjustments
    ENABLE ALWAYS TRIGGER billing_adjustments_authority_guard;

CREATE TRIGGER billing_discount_redemptions_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_discount_redemptions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_discount_redemptions
    ENABLE ALWAYS TRIGGER billing_discount_redemptions_authority_guard;

CREATE TRIGGER billing_entitlement_decisions_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_entitlement_decisions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_entitlement_decisions
    ENABLE ALWAYS TRIGGER billing_entitlement_decisions_authority_guard;

CREATE TRIGGER billing_invoice_lines_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_invoice_lines
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_invoice_lines
    ENABLE ALWAYS TRIGGER billing_invoice_lines_authority_guard;

CREATE TRIGGER billing_invoice_finalizations_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_invoice_finalizations
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_invoice_finalizations
    ENABLE ALWAYS TRIGGER billing_invoice_finalizations_authority_guard;

CREATE TRIGGER billing_invoices_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_invoices
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_invoices
    ENABLE ALWAYS TRIGGER billing_invoices_authority_guard;

CREATE TRIGGER billing_meter_periods_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_meter_periods
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_meter_periods
    ENABLE ALWAYS TRIGGER billing_meter_periods_authority_guard;

CREATE TRIGGER billing_provider_commands_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_provider_commands
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_provider_commands
    ENABLE ALWAYS TRIGGER billing_provider_commands_authority_guard;

CREATE TRIGGER billing_provider_events_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_provider_events
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_provider_events
    ENABLE ALWAYS TRIGGER billing_provider_events_authority_guard;

CREATE TRIGGER billing_provider_objects_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_provider_objects
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_provider_objects
    ENABLE ALWAYS TRIGGER billing_provider_objects_authority_guard;

CREATE TRIGGER billing_settlements_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_settlements
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_settlements
    ENABLE ALWAYS TRIGGER billing_settlements_authority_guard;

CREATE TRIGGER billing_subscription_changes_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_changes
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_subscription_changes
    ENABLE ALWAYS TRIGGER billing_subscription_changes_authority_guard;

CREATE TRIGGER billing_subscription_overrides_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_overrides
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_subscription_overrides
    ENABLE ALWAYS TRIGGER billing_subscription_overrides_authority_guard;

CREATE TRIGGER billing_subscription_operations_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_operations
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_subscription_operations
    ENABLE ALWAYS TRIGGER billing_subscription_operations_authority_guard;

CREATE TRIGGER billing_subscriptions_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscriptions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_subscriptions
    ENABLE ALWAYS TRIGGER billing_subscriptions_authority_guard;

CREATE TRIGGER billing_tenant_base_offerings_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_tenant_base_offerings
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_tenant_base_offerings
    ENABLE ALWAYS TRIGGER billing_tenant_base_offerings_authority_guard;

CREATE TRIGGER billing_trial_claims_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_trial_claims
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_trial_claims
    ENABLE ALWAYS TRIGGER billing_trial_claims_authority_guard;

CREATE TRIGGER billing_usage_entries_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_usage_entries
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_usage_entries
    ENABLE ALWAYS TRIGGER billing_usage_entries_authority_guard;

CREATE TRIGGER billing_usage_reservations_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_usage_reservations
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_usage_reservations
    ENABLE ALWAYS TRIGGER billing_usage_reservations_authority_guard;

CREATE TRIGGER billing_usage_totals_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_usage_totals
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_usage_totals
    ENABLE ALWAYS TRIGGER billing_usage_totals_authority_guard;

CREATE TRIGGER billing_webhook_receipts_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_webhook_receipts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_webhook_receipts
    ENABLE ALWAYS TRIGGER billing_webhook_receipts_authority_guard;
