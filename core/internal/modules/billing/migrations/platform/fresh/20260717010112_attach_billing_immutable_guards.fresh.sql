-- +goose Up
CREATE TRIGGER billing_adjustments_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_adjustments
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_adjustments
    ENABLE ALWAYS TRIGGER billing_adjustments_immutable_guard;

CREATE TRIGGER billing_campaign_discounts_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_campaign_discounts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_campaign_discounts
    ENABLE ALWAYS TRIGGER billing_campaign_discounts_immutable_guard;

CREATE TRIGGER billing_campaign_versions_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_campaign_versions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_campaign_versions
    ENABLE ALWAYS TRIGGER billing_campaign_versions_immutable_guard;

CREATE TRIGGER billing_catalog_operations_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_catalog_operations
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_catalog_operations
    ENABLE ALWAYS TRIGGER billing_catalog_operations_immutable_guard;

CREATE TRIGGER billing_currency_exponents_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_currency_exponents
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_currency_exponents
    ENABLE ALWAYS TRIGGER billing_currency_exponents_immutable_guard;

CREATE TRIGGER billing_currency_registry_versions_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_currency_registry_versions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_currency_registry_versions
    ENABLE ALWAYS TRIGGER billing_currency_registry_versions_immutable_guard;

CREATE TRIGGER billing_discount_offerings_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_discount_offerings
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_discount_offerings
    ENABLE ALWAYS TRIGGER billing_discount_offerings_immutable_guard;

CREATE TRIGGER billing_discount_redemptions_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_discount_redemptions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_discount_redemptions
    ENABLE ALWAYS TRIGGER billing_discount_redemptions_immutable_guard;

CREATE TRIGGER billing_discount_versions_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_discount_versions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_discount_versions
    ENABLE ALWAYS TRIGGER billing_discount_versions_immutable_guard;

CREATE TRIGGER billing_entitlement_decisions_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_entitlement_decisions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_entitlement_decisions
    ENABLE ALWAYS TRIGGER billing_entitlement_decisions_immutable_guard;

CREATE TRIGGER billing_entitlement_definitions_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_entitlement_definitions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_entitlement_definitions
    ENABLE ALWAYS TRIGGER billing_entitlement_definitions_immutable_guard;

CREATE TRIGGER billing_entitlement_registry_versions_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_entitlement_registry_versions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_entitlement_registry_versions
    ENABLE ALWAYS TRIGGER billing_entitlement_registry_versions_immutable_guard;

CREATE TRIGGER billing_offering_entitlements_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_offering_entitlements
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_offering_entitlements
    ENABLE ALWAYS TRIGGER billing_offering_entitlements_immutable_guard;

CREATE TRIGGER billing_invoice_finalizations_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_invoice_finalizations
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_invoice_finalizations
    ENABLE ALWAYS TRIGGER billing_invoice_finalizations_immutable_guard;

CREATE TRIGGER billing_offering_versions_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_offering_versions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_offering_versions
    ENABLE ALWAYS TRIGGER billing_offering_versions_immutable_guard;

CREATE TRIGGER billing_price_versions_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_price_versions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_price_versions
    ENABLE ALWAYS TRIGGER billing_price_versions_immutable_guard;

CREATE TRIGGER billing_provider_events_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_provider_events
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_provider_events
    ENABLE ALWAYS TRIGGER billing_provider_events_immutable_guard;

CREATE TRIGGER billing_subscription_overrides_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_overrides
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_subscription_overrides
    ENABLE ALWAYS TRIGGER billing_subscription_overrides_immutable_guard;

CREATE TRIGGER billing_subscription_operations_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_operations
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_subscription_operations
    ENABLE ALWAYS TRIGGER billing_subscription_operations_immutable_guard;

CREATE TRIGGER billing_trial_claims_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_trial_claims
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_trial_claims
    ENABLE ALWAYS TRIGGER billing_trial_claims_immutable_guard;

CREATE TRIGGER billing_usage_entries_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_usage_entries
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_usage_entries
    ENABLE ALWAYS TRIGGER billing_usage_entries_immutable_guard;
