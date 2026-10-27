-- +goose Up
CREATE TRIGGER billing_campaign_discounts_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_campaign_discounts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_campaign_discounts
    ENABLE ALWAYS TRIGGER billing_campaign_discounts_authority_guard;

CREATE TRIGGER billing_campaign_versions_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_campaign_versions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_campaign_versions
    ENABLE ALWAYS TRIGGER billing_campaign_versions_authority_guard;

CREATE TRIGGER billing_campaigns_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_campaigns
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_campaigns
    ENABLE ALWAYS TRIGGER billing_campaigns_authority_guard;

CREATE TRIGGER billing_catalog_operations_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_catalog_operations
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_catalog_operations
    ENABLE ALWAYS TRIGGER billing_catalog_operations_authority_guard;

CREATE TRIGGER billing_currency_exponents_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_currency_exponents
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_currency_exponents
    ENABLE ALWAYS TRIGGER billing_currency_exponents_authority_guard;

CREATE TRIGGER billing_currency_registry_versions_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_currency_registry_versions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_currency_registry_versions
    ENABLE ALWAYS TRIGGER billing_currency_registry_versions_authority_guard;

CREATE TRIGGER billing_discount_offerings_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_discount_offerings
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_discount_offerings
    ENABLE ALWAYS TRIGGER billing_discount_offerings_authority_guard;

CREATE TRIGGER billing_discount_versions_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_discount_versions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_discount_versions
    ENABLE ALWAYS TRIGGER billing_discount_versions_authority_guard;

CREATE TRIGGER billing_discounts_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_discounts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_discounts
    ENABLE ALWAYS TRIGGER billing_discounts_authority_guard;

CREATE TRIGGER billing_entitlement_definitions_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_entitlement_definitions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_entitlement_definitions
    ENABLE ALWAYS TRIGGER billing_entitlement_definitions_authority_guard;

CREATE TRIGGER billing_entitlement_registry_versions_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_entitlement_registry_versions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_entitlement_registry_versions
    ENABLE ALWAYS TRIGGER billing_entitlement_registry_versions_authority_guard;

CREATE TRIGGER billing_offering_entitlements_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_offering_entitlements
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_offering_entitlements
    ENABLE ALWAYS TRIGGER billing_offering_entitlements_authority_guard;

CREATE TRIGGER billing_offering_versions_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_offering_versions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_offering_versions
    ENABLE ALWAYS TRIGGER billing_offering_versions_authority_guard;

CREATE TRIGGER billing_offerings_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_offerings
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_offerings
    ENABLE ALWAYS TRIGGER billing_offerings_authority_guard;

CREATE TRIGGER billing_price_versions_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_price_versions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_price_versions
    ENABLE ALWAYS TRIGGER billing_price_versions_authority_guard;

CREATE TRIGGER billing_provider_accounts_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_provider_accounts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_row_v1();
ALTER TABLE platform.billing_provider_accounts
    ENABLE ALWAYS TRIGGER billing_provider_accounts_authority_guard;
