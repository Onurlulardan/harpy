-- +goose Up
-- Migration version: 20260717010211 UTC
ALTER TABLE platform.billing_adjustments
    DISABLE TRIGGER billing_adjustments_immutable_guard;
CREATE TRIGGER billing_adjustments_retention_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_adjustments
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_retention_v1();
ALTER TABLE platform.billing_adjustments
    ENABLE ALWAYS TRIGGER billing_adjustments_retention_immutable_guard;

ALTER TABLE platform.billing_discount_redemptions
    DISABLE TRIGGER billing_discount_redemptions_immutable_guard;
CREATE TRIGGER billing_discount_redemptions_retention_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_discount_redemptions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_retention_v1();
ALTER TABLE platform.billing_discount_redemptions
    ENABLE ALWAYS TRIGGER billing_discount_redemptions_retention_immutable_guard;

ALTER TABLE platform.billing_entitlement_decisions
    DISABLE TRIGGER billing_entitlement_decisions_immutable_guard;
CREATE TRIGGER billing_entitlement_decisions_retention_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_entitlement_decisions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_retention_v1();
ALTER TABLE platform.billing_entitlement_decisions
    ENABLE ALWAYS TRIGGER billing_entitlement_decisions_retention_immutable_guard;

ALTER TABLE platform.billing_invoice_finalizations
    DISABLE TRIGGER billing_invoice_finalizations_immutable_guard;
CREATE TRIGGER billing_invoice_finalizations_retention_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_invoice_finalizations
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_retention_v1();
ALTER TABLE platform.billing_invoice_finalizations
    ENABLE ALWAYS TRIGGER billing_invoice_finalizations_retention_immutable_guard;

ALTER TABLE platform.billing_subscription_overrides
    DISABLE TRIGGER billing_subscription_overrides_immutable_guard;
CREATE TRIGGER billing_subscription_overrides_retention_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_overrides
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_retention_v1();
ALTER TABLE platform.billing_subscription_overrides
    ENABLE ALWAYS TRIGGER billing_subscription_overrides_retention_immutable_guard;

ALTER TABLE platform.billing_subscription_operations
    DISABLE TRIGGER billing_subscription_operations_immutable_guard;
CREATE TRIGGER billing_subscription_operations_retention_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_operations
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_retention_v1();
ALTER TABLE platform.billing_subscription_operations
    ENABLE ALWAYS TRIGGER billing_subscription_operations_retention_immutable_guard;

ALTER TABLE platform.billing_trial_claims
    DISABLE TRIGGER billing_trial_claims_immutable_guard;
CREATE TRIGGER billing_trial_claims_retention_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_trial_claims
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_retention_v1();
ALTER TABLE platform.billing_trial_claims
    ENABLE ALWAYS TRIGGER billing_trial_claims_retention_immutable_guard;

ALTER TABLE platform.billing_usage_entries
    DISABLE TRIGGER billing_usage_entries_immutable_guard;
CREATE TRIGGER billing_usage_entries_retention_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_usage_entries
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_retention_v1();
ALTER TABLE platform.billing_usage_entries
    ENABLE ALWAYS TRIGGER billing_usage_entries_retention_immutable_guard;

ALTER TABLE platform.billing_subscription_audit
    DISABLE TRIGGER billing_subscription_audit_immutable_guard;
CREATE TRIGGER billing_subscription_audit_retention_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_audit
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_retention_v1();
ALTER TABLE platform.billing_subscription_audit
    ENABLE ALWAYS TRIGGER billing_subscription_audit_retention_immutable_guard;

ALTER TABLE platform.billing_subscription_boundary_jobs
    DISABLE TRIGGER billing_subscription_boundary_jobs_immutable_guard;
CREATE TRIGGER billing_subscription_boundary_jobs_retention_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_boundary_jobs
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_retention_v1();
ALTER TABLE platform.billing_subscription_boundary_jobs
    ENABLE ALWAYS TRIGGER billing_subscription_boundary_jobs_retention_immutable_guard;

ALTER TABLE platform.billing_subscription_boundary_evidence
    DISABLE TRIGGER billing_subscription_boundary_evidence_immutable_guard;
CREATE TRIGGER billing_subscription_boundary_evidence_retention_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_boundary_evidence
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_retention_v1();
ALTER TABLE platform.billing_subscription_boundary_evidence
    ENABLE ALWAYS TRIGGER billing_subscription_boundary_evidence_retention_immutable_guard;

ALTER TABLE platform.billing_usage_totals
    DISABLE TRIGGER billing_usage_totals_revision_guard;
CREATE TRIGGER billing_usage_totals_retention_revision_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_usage_totals
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_usage_total_retention_v1();
ALTER TABLE platform.billing_usage_totals
    ENABLE ALWAYS TRIGGER billing_usage_totals_retention_revision_guard;

ALTER TABLE platform.billing_usage_reservations
    DISABLE TRIGGER billing_usage_reservations_transition_guard;
CREATE TRIGGER billing_usage_reservations_retention_transition_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_usage_reservations
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_usage_reservation_retention_v1();
ALTER TABLE platform.billing_usage_reservations
    ENABLE ALWAYS TRIGGER billing_usage_reservations_retention_transition_guard;

ALTER TABLE platform.billing_invoices
    DISABLE TRIGGER billing_invoices_finalization_guard;
CREATE TRIGGER billing_invoices_retention_finalization_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_invoices
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_invoice_retention_v1();
ALTER TABLE platform.billing_invoices
    ENABLE ALWAYS TRIGGER billing_invoices_retention_finalization_guard;

ALTER TABLE platform.billing_invoice_lines
    DISABLE TRIGGER billing_invoice_lines_finalization_guard;
CREATE TRIGGER billing_invoice_lines_retention_finalization_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_invoice_lines
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_invoice_line_retention_v1();
ALTER TABLE platform.billing_invoice_lines
    ENABLE ALWAYS TRIGGER billing_invoice_lines_retention_finalization_guard;

ALTER TABLE platform.billing_adjustments
    DISABLE TRIGGER billing_adjustments_finalization_guard;
CREATE TRIGGER billing_adjustments_retention_finalization_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_adjustments
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_invoice_line_retention_v1();
ALTER TABLE platform.billing_adjustments
    ENABLE ALWAYS TRIGGER billing_adjustments_retention_finalization_guard;

ALTER TABLE platform.billing_settlements
    DISABLE TRIGGER billing_settlements_state_guard;
CREATE TRIGGER billing_settlements_retention_state_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_settlements
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_settlement_retention_v1();
ALTER TABLE platform.billing_settlements
    ENABLE ALWAYS TRIGGER billing_settlements_retention_state_guard;

ALTER TABLE platform.billing_subscriptions
    DISABLE TRIGGER billing_subscriptions_transition_guard;
CREATE TRIGGER billing_subscriptions_retention_transition_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscriptions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_subscription_retention_v1();
ALTER TABLE platform.billing_subscriptions
    ENABLE ALWAYS TRIGGER billing_subscriptions_retention_transition_guard;
