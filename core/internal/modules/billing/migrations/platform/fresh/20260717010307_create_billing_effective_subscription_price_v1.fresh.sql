-- +goose Up
-- Migration version: 20260717010307 UTC
CREATE TABLE platform.billing_v1_subscription_wire_archive (
    subscription_id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    source_subscription_id uuid NOT NULL,
    source_updated_at timestamptz NOT NULL,
    trial_started_at timestamptz,
    cancelled_at timestamptz,
    cancellation_reason text NOT NULL,
    adoption_operation_id uuid NOT NULL,
    adoption_run_id uuid NOT NULL,
    archived_at timestamptz NOT NULL,
    CONSTRAINT billing_v1_subscription_wire_archive_pkey
        PRIMARY KEY (subscription_id),
    CONSTRAINT billing_v1_subscription_wire_archive_source_key
        UNIQUE (tenant_id, source_subscription_id),
    CONSTRAINT billing_v1_subscription_wire_archive_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_v1_subscription_wire_archive_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_v1_subscription_wire_archive_run_fk FOREIGN KEY (
        adoption_operation_id, adoption_run_id
    ) REFERENCES platform.billing_v1_adoption_runs (operation_id, run_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);
CREATE TRIGGER billing_v1_subscription_wire_archive_authority_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_v1_subscription_wire_archive
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_v1_subscription_wire_archive ENABLE ALWAYS TRIGGER
    billing_v1_subscription_wire_archive_authority_guard;
CREATE TRIGGER billing_v1_subscription_wire_archive_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_v1_subscription_wire_archive
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_v1_subscription_wire_archive ENABLE ALWAYS TRIGGER
    billing_v1_subscription_wire_archive_immutable_guard;

-- +goose StatementBegin
CREATE FUNCTION platform.billing_effective_subscription_price_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_subscription_id uuid,
    p_at timestamptz
)
RETURNS TABLE (
    effective_minor_units bigint,
    effective_currency text,
    effective_registry_version bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_state text;
    v_price_version_id uuid;
    v_applied_kind text;
    v_applied_currency text;
    v_applied_registry bigint;
    v_applied_fixed bigint;
    v_applied_basis_points integer;
    v_quotient bigint;
    v_remainder bigint;
    v_discount bigint;
BEGIN
    IF p_authority_id IS NULL OR p_tenant_id IS NULL
       OR p_subscription_id IS NULL OR p_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing effective price input is invalid';
    END IF;
    PERFORM platform.billing_assert_retained_tenant_authority_v1(
        p_authority_id, p_tenant_id
    );
    SELECT subscription.state, subscription.price_version_id,
           price.minor_units, pg_catalog.btrim(price.currency::text),
           price.currency_registry_version
    INTO STRICT v_state, v_price_version_id, effective_minor_units,
        effective_currency, effective_registry_version
    FROM platform.billing_subscriptions AS subscription
    JOIN platform.billing_price_versions AS price
      ON price.id = subscription.price_version_id
     AND price.offering_version_id = subscription.offering_version_id
    WHERE subscription.authority_id = p_authority_id
      AND subscription.tenant_id = p_tenant_id
      AND subscription.id = p_subscription_id;
    IF v_state = 'trialing' THEN
        effective_minor_units := 0;
        RETURN NEXT;
        RETURN;
    END IF;
    FOR v_applied_kind, v_applied_currency, v_applied_registry,
        v_applied_fixed, v_applied_basis_points IN
        SELECT redemption.applied_kind,
               pg_catalog.btrim(redemption.applied_currency::text),
               redemption.applied_currency_registry_version,
               redemption.applied_fixed_minor_units,
               redemption.applied_basis_points
        FROM platform.billing_discount_redemptions AS redemption
        WHERE redemption.authority_id = p_authority_id
          AND redemption.tenant_id = p_tenant_id
          AND redemption.subscription_id = p_subscription_id
          AND redemption.price_version_id = v_price_version_id
          AND redemption.redeemed_at <= p_at
          AND (redemption.relationship_ended_at IS NULL
               OR redemption.relationship_ended_at > p_at)
        ORDER BY redemption.applied_priority,
                 redemption.discount_version_id
    LOOP
        IF v_applied_kind = 'fixed' THEN
            IF v_applied_currency IS DISTINCT FROM effective_currency
               OR v_applied_registry
                    IS DISTINCT FROM effective_registry_version THEN
                RAISE EXCEPTION USING ERRCODE = '23514',
                    MESSAGE = 'billing discount currency invariant failed';
            END IF;
            v_discount := v_applied_fixed;
        ELSE
            SELECT pg_catalog.floor(
                       effective_minor_units::numeric
                       * v_applied_basis_points::numeric / 10000
                   )::bigint,
                   pg_catalog.mod(
                       effective_minor_units::numeric
                       * v_applied_basis_points::numeric, 10000
                   )::bigint
            INTO v_quotient, v_remainder;
            IF v_remainder > 5000
               OR (v_remainder = 5000 AND pg_catalog.mod(v_quotient, 2) = 1) THEN
                v_quotient := v_quotient + 1;
            END IF;
            v_discount := v_quotient;
        END IF;
        IF v_discount >= effective_minor_units THEN
            effective_minor_units := 0;
        ELSE
            effective_minor_units := effective_minor_units - v_discount;
        END IF;
    END LOOP;
    RETURN NEXT;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_effective_subscription_price_v1(
    uuid, uuid, uuid, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_effective_subscription_price_v1(
    uuid, uuid, uuid, timestamptz
) TO forgecrud_runtime;
-- +goose StatementEnd
