-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_invoice_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    selected_line_count bigint;
    selected_gross bigint;
    selected_discount bigint;
    selected_tax bigint;
    selected_debit bigint;
    selected_credit bigint;
    selected_total bigint;
    settlement_total bigint;
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        IF NEW.revision IS DISTINCT FROM 1
           OR (SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime'
               AND NEW.state IS DISTINCT FROM 'draft') THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'billing invoice initial state is invalid';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN
        IF OLD.state IS DISTINCT FROM 'draft' THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'billing finalized invoice cannot be deleted';
        END IF;
        RETURN OLD;
    END IF;
    IF NEW.id IS DISTINCT FROM OLD.id
       OR NEW.authority_id IS DISTINCT FROM OLD.authority_id
       OR NEW.tenant_id IS DISTINCT FROM OLD.tenant_id
       OR NEW.subscription_id IS DISTINCT FROM OLD.subscription_id
       OR NEW.invoice_number IS DISTINCT FROM OLD.invoice_number
       OR NEW.currency IS DISTINCT FROM OLD.currency
       OR NEW.currency_registry_version IS DISTINCT FROM OLD.currency_registry_version
       OR NEW.currency_eligible IS DISTINCT FROM OLD.currency_eligible
       OR NEW.money_policy_version IS DISTINCT FROM OLD.money_policy_version
       OR NEW.policy_sha256 IS DISTINCT FROM OLD.policy_sha256
       OR NEW.period_start IS DISTINCT FROM OLD.period_start
       OR NEW.period_end IS DISTINCT FROM OLD.period_end
       OR NEW.created_at IS DISTINCT FROM OLD.created_at
       OR NEW.revision IS DISTINCT FROM OLD.revision + 1
       OR NEW.updated_at <= OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing invoice identity or revision is invalid';
    END IF;
    IF OLD.state IS NOT DISTINCT FROM 'finalized' THEN
        IF NEW.state IS DISTINCT FROM OLD.state
           OR NEW.gross_minor_units IS DISTINCT FROM OLD.gross_minor_units
           OR NEW.discount_minor_units IS DISTINCT FROM OLD.discount_minor_units
           OR NEW.tax_minor_units IS DISTINCT FROM OLD.tax_minor_units
           OR NEW.debit_adjustment_minor_units IS DISTINCT FROM
               OLD.debit_adjustment_minor_units
           OR NEW.credit_adjustment_minor_units IS DISTINCT FROM
               OLD.credit_adjustment_minor_units
           OR NEW.invoice_total_minor_units IS DISTINCT FROM
               OLD.invoice_total_minor_units
           OR NEW.finalized_at IS DISTINCT FROM OLD.finalized_at THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'billing finalized invoice snapshot is immutable';
        END IF;
        SELECT COALESCE(pg_catalog.sum(
            CASE settlement_kind WHEN 'paid' THEN minor_units ELSE -minor_units END
        ), 0)::bigint INTO settlement_total
        FROM platform.billing_settlements
        WHERE tenant_id = NEW.tenant_id AND invoice_id = NEW.id;
        IF settlement_total IS DISTINCT FROM NEW.settled_minor_units::numeric THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'billing invoice settlement total is inconsistent';
        END IF;
        RETURN NEW;
    END IF;
    IF NEW.state IS DISTINCT FROM 'draft'
       AND NEW.state IS DISTINCT FROM 'finalized' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing invoice state transition is forbidden';
    END IF;
    IF NEW.state IS NOT DISTINCT FROM 'finalized' THEN
        SELECT pg_catalog.count(*),
               COALESCE(pg_catalog.sum(gross_minor_units), 0)::bigint,
               COALESCE(pg_catalog.sum(discount_minor_units), 0)::bigint,
               COALESCE(pg_catalog.sum(tax_minor_units), 0)::bigint,
               COALESCE(pg_catalog.sum(debit_adjustment_minor_units), 0)::bigint,
               COALESCE(pg_catalog.sum(credit_adjustment_minor_units), 0)::bigint,
               COALESCE(pg_catalog.sum(line_total_minor_units), 0)::bigint
        INTO selected_line_count, selected_gross, selected_discount,
             selected_tax, selected_debit, selected_credit, selected_total
        FROM platform.billing_invoice_lines
        WHERE tenant_id = NEW.tenant_id AND invoice_id = NEW.id;
        IF selected_line_count = 0 OR selected_gross <> NEW.gross_minor_units
           OR selected_discount <> NEW.discount_minor_units
           OR selected_tax <> NEW.tax_minor_units
           OR selected_debit <> NEW.debit_adjustment_minor_units
           OR selected_credit <> NEW.credit_adjustment_minor_units
           OR selected_total <> NEW.invoice_total_minor_units THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'billing invoice lines do not match final snapshot';
        END IF;
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_invoice_v1() FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_invoice_line_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE selected_tenant uuid; selected_invoice uuid;
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN
        selected_tenant := OLD.tenant_id; selected_invoice := OLD.invoice_id;
    ELSE
        selected_tenant := NEW.tenant_id; selected_invoice := NEW.invoice_id;
    END IF;
    PERFORM 1 FROM platform.billing_invoices
    WHERE tenant_id = selected_tenant AND id = selected_invoice AND state = 'draft'
    FOR KEY SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing finalized invoice lines cannot change';
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_invoice_line_v1() FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_settlement_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'INSERT' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing settlement fact is immutable';
    END IF;
    PERFORM 1 FROM platform.billing_invoices
    WHERE tenant_id = NEW.tenant_id AND id = NEW.invoice_id AND state = 'finalized'
    FOR KEY SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing settlement requires a finalized invoice';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_settlement_v1() FROM PUBLIC;

CREATE TRIGGER billing_invoices_finalization_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_invoices
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_invoice_v1();
ALTER TABLE platform.billing_invoices
    ENABLE ALWAYS TRIGGER billing_invoices_finalization_guard;
CREATE TRIGGER billing_invoice_lines_finalization_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_invoice_lines
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_invoice_line_v1();
ALTER TABLE platform.billing_invoice_lines
    ENABLE ALWAYS TRIGGER billing_invoice_lines_finalization_guard;
CREATE TRIGGER billing_adjustments_finalization_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_adjustments
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_invoice_line_v1();
ALTER TABLE platform.billing_adjustments
    ENABLE ALWAYS TRIGGER billing_adjustments_finalization_guard;
CREATE TRIGGER billing_settlements_state_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_settlements
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_settlement_v1();
ALTER TABLE platform.billing_settlements
    ENABLE ALWAYS TRIGGER billing_settlements_state_guard;
-- +goose StatementEnd
