-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_revenue_projection_id_v1(
    p_subscription_id uuid
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    value text;
BEGIN
	IF p_subscription_id IS NULL THEN
		RAISE EXCEPTION USING ERRCODE = '22023',
			MESSAGE = 'billing revenue projection identity is invalid';
	END IF;
    value := pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        'billing-revenue-v1:' || p_subscription_id::text, 'UTF8'
    )), 'hex');
    RETURN (
        pg_catalog.substr(value, 1, 8) || '-' ||
        pg_catalog.substr(value, 9, 4) || '-5' ||
        pg_catalog.substr(value, 14, 3) || '-8' ||
        pg_catalog.substr(value, 18, 3) || '-' ||
        pg_catalog.substr(value, 21, 12)
    )::uuid;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.billing_revenue_projection_id_v1(uuid) FROM PUBLIC;

CREATE FUNCTION platform.billing_revenue_monthly_v1(
    p_minor_units bigint,
    p_cadence text
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    quotient bigint;
    remainder bigint;
BEGIN
    IF p_minor_units IS NULL OR p_cadence IS NULL OR p_minor_units < 0
       OR p_cadence NOT IN ('monthly', 'yearly') THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing revenue normalization input is invalid';
    END IF;
    IF p_cadence = 'monthly' THEN
        RETURN p_minor_units;
    END IF;
    quotient := p_minor_units / 12;
    remainder := p_minor_units % 12;
    IF remainder > 6 OR (remainder = 6 AND quotient % 2 = 1) THEN
        quotient := quotient + 1;
    END IF;
    RETURN quotient;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.billing_revenue_monthly_v1(bigint, text) FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_revenue_projection_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY INVOKER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing revenue projection writer is invalid';
    END IF;
    IF TG_OP IS DISTINCT FROM 'DELETE' THEN
        RETURN NEW;
    END IF;
    RETURN OLD;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.billing_guard_revenue_projection_v1() FROM PUBLIC;

CREATE TRIGGER billing_revenue_projection_writer_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_revenue_projection
FOR EACH ROW EXECUTE FUNCTION
    platform.billing_guard_revenue_projection_v1();
ALTER TABLE platform.billing_revenue_projection
    ENABLE ALWAYS TRIGGER billing_revenue_projection_writer_guard;

CREATE TRIGGER billing_revenue_rebuilds_writer_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_revenue_rebuilds
FOR EACH ROW EXECUTE FUNCTION
    platform.billing_guard_revenue_projection_v1();
ALTER TABLE platform.billing_revenue_rebuilds
    ENABLE ALWAYS TRIGGER billing_revenue_rebuilds_writer_guard;
-- +goose StatementEnd
