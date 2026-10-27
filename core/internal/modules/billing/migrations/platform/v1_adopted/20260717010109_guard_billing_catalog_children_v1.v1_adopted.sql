-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_offering_children_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    selected_price_count bigint;
    selected_grant_count bigint;
BEGIN
    SELECT pg_catalog.count(*) INTO selected_price_count
    FROM platform.billing_price_versions
    WHERE authority_id = NEW.authority_id
      AND offering_version_id = NEW.id;
    SELECT pg_catalog.count(*) INTO selected_grant_count
    FROM platform.billing_offering_entitlements
    WHERE authority_id = NEW.authority_id
      AND offering_version_id = NEW.id;
    IF selected_price_count IS DISTINCT FROM NEW.expected_price_count::bigint
       OR selected_grant_count IS DISTINCT FROM NEW.expected_grant_count::bigint THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'billing offering publication children are incomplete';
    END IF;
    RETURN NEW;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_guard_offering_children_v1() FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_discount_children_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    selected_offering_count bigint;
BEGIN
    SELECT pg_catalog.count(*) INTO selected_offering_count
    FROM platform.billing_discount_offerings
    WHERE authority_id = NEW.authority_id
      AND discount_version_id = NEW.id;
    IF selected_offering_count IS DISTINCT FROM NEW.expected_offering_count::bigint THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'billing discount publication children are incomplete';
    END IF;
    RETURN NEW;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_guard_discount_children_v1() FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_campaign_children_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    selected_discount_count bigint;
BEGIN
    SELECT pg_catalog.count(*) INTO selected_discount_count
    FROM platform.billing_campaign_discounts
    WHERE authority_id = NEW.authority_id
      AND campaign_version_id = NEW.id;
    IF selected_discount_count IS DISTINCT FROM NEW.expected_discount_count::bigint THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'billing campaign publication children are incomplete';
    END IF;
    RETURN NEW;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_guard_campaign_children_v1() FROM PUBLIC;

CREATE CONSTRAINT TRIGGER billing_offering_versions_children_guard
AFTER INSERT ON platform.billing_offering_versions
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_offering_children_v1();
ALTER TABLE platform.billing_offering_versions
    ENABLE ALWAYS TRIGGER billing_offering_versions_children_guard;

CREATE CONSTRAINT TRIGGER billing_discount_versions_children_guard
AFTER INSERT ON platform.billing_discount_versions
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_discount_children_v1();
ALTER TABLE platform.billing_discount_versions
    ENABLE ALWAYS TRIGGER billing_discount_versions_children_guard;

CREATE CONSTRAINT TRIGGER billing_campaign_versions_children_guard
AFTER INSERT ON platform.billing_campaign_versions
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_campaign_children_v1();
ALTER TABLE platform.billing_campaign_versions
    ENABLE ALWAYS TRIGGER billing_campaign_versions_children_guard;
-- +goose StatementEnd

