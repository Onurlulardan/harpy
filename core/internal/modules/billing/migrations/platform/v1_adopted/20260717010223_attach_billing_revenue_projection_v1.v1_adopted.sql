-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_maintain_revenue_projection_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN
        PERFORM platform.billing_refresh_revenue_projection_v1(OLD.id);
        RETURN OLD;
    END IF;
    PERFORM platform.billing_refresh_revenue_projection_v1(NEW.id);
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.billing_maintain_revenue_projection_v1() FROM PUBLIC;

CREATE TRIGGER billing_subscriptions_revenue_projection
AFTER INSERT OR UPDATE OR DELETE
ON platform.billing_subscriptions
FOR EACH ROW EXECUTE FUNCTION
    platform.billing_maintain_revenue_projection_v1();
ALTER TABLE platform.billing_subscriptions
    ENABLE ALWAYS TRIGGER billing_subscriptions_revenue_projection;
-- +goose StatementEnd

