-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_v1_base_adoption_assert_v1(
    p_operation_id uuid,p_environment_id uuid,p_tenant_id uuid,
    p_evidence_kind text,p_offering_key text,p_offering_version_id uuid,
    p_offering_revision bigint,p_offering_sha256 text,p_offering_effective_at timestamptz,
    p_source_sequence bigint,p_source_sha256 text,p_subscription_sha256 text,p_plan_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.billing_v1_base_adoption_receipts%ROWTYPE;
    v_assignment_id uuid;
    v_assignment_count integer;
    v_purpose text;
    v_source_sequence bigint;
    v_source_sha256 text;
    v_subscription_sha256 text;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    v_source_sequence := NULLIF(p_source_sequence,0);
    v_source_sha256 := NULLIF(p_source_sha256,'');
    v_subscription_sha256 := NULLIF(p_subscription_sha256,'');
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name',true)
            !~ '^forgecrud-cutover/[0-9a-f-]{36}$'
       OR p_operation_id IS NULL OR p_environment_id IS NULL OR p_tenant_id IS NULL
       OR p_evidence_kind NOT IN ('license_derived','saas_control','saas_free','saas_paid')
       OR p_offering_key !~ '^[a-z][a-z0-9._-]{0,126}[a-z0-9]$'
       OR p_offering_version_id IS NULL OR p_offering_revision<1
       OR p_offering_sha256 !~ '^[0-9a-f]{64}$'
       OR NOT pg_catalog.isfinite(p_offering_effective_at)
       OR p_plan_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='V1 Billing base adoption denied';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended('v1-billing-base|'||p_tenant_id::text,0));
    SELECT * INTO v_existing FROM platform.billing_v1_base_adoption_receipts
     WHERE tenant_id=p_tenant_id FOR SHARE;
    IF FOUND THEN
        IF v_existing.operation_id IS DISTINCT FROM p_operation_id
           OR v_existing.environment_id IS DISTINCT FROM p_environment_id
           OR v_existing.evidence_kind IS DISTINCT FROM p_evidence_kind
           OR v_existing.offering_key IS DISTINCT FROM p_offering_key
           OR v_existing.offering_version_id IS DISTINCT FROM p_offering_version_id
           OR v_existing.offering_revision IS DISTINCT FROM p_offering_revision
           OR v_existing.offering_sha256 IS DISTINCT FROM p_offering_sha256
           OR v_existing.offering_effective_at IS DISTINCT FROM p_offering_effective_at
           OR v_existing.source_sequence IS DISTINCT FROM v_source_sequence
           OR v_existing.source_sha256 IS DISTINCT FROM v_source_sha256
           OR v_existing.subscription_evidence_sha256 IS DISTINCT FROM v_subscription_sha256
           OR v_existing.plan_sha256 IS DISTINCT FROM p_plan_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='V1 Billing base adoption replay conflicts';
        END IF;
        RETURN true;
    END IF;
    SELECT binding.purpose INTO v_purpose
      FROM platform.tenancy_tenant_bindings AS binding
     WHERE binding.environment_id=p_environment_id AND binding.tenant_id=p_tenant_id
     FOR SHARE;
    IF (p_evidence_kind='license_derived' AND v_purpose<>'onpremise_primary')
       OR (p_evidence_kind='saas_control' AND v_purpose<>'saas_control')
       OR (p_evidence_kind IN ('saas_free','saas_paid') AND v_purpose<>'saas_customer') THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='V1 Billing evidence conflicts with tenant purpose';
    END IF;
    PERFORM 1 FROM platform.billing_tenant_base_offerings AS base
     WHERE base.tenant_id=p_tenant_id AND base.effective_end IS NULL
     FOR SHARE;
    GET DIAGNOSTICS v_assignment_count = ROW_COUNT;
    IF v_assignment_count IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='V1 Billing base assignment inventory is ambiguous';
    END IF;
    SELECT base.id INTO v_assignment_id
      FROM platform.billing_tenant_base_offerings AS base
      JOIN platform.billing_offering_versions AS version ON version.id=base.offering_version_id
      JOIN platform.billing_offerings AS offering ON offering.id=version.offering_id
     WHERE base.tenant_id=p_tenant_id AND base.effective_end IS NULL
       AND base.offering_version_id=p_offering_version_id
       AND version.version=p_offering_revision
       AND offering.offering_key=p_offering_key AND offering.lifecycle='active'
     FOR SHARE OF base,version,offering;
    IF v_assignment_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='exact V2 base assignment is unavailable';
    END IF;
    IF p_evidence_kind='license_derived' THEN
        PERFORM 1 FROM platform.billing_license_offering_publications AS publication
         WHERE publication.environment_id=p_environment_id
           AND publication.offering_version_id=p_offering_version_id
           AND publication.revision=p_offering_revision
           AND publication.offering_sha256=p_offering_sha256
           AND publication.effective_at=p_offering_effective_at
           AND publication.snapshot_sequence=v_source_sequence
           AND publication.snapshot_sha256=v_source_sha256 FOR SHARE;
    ELSIF p_evidence_kind='saas_control' THEN
        PERFORM 1 FROM platform.billing_saas_control_offering_publications AS publication
         WHERE publication.offering_key='saas_platform_control_base_v1'
           AND p_offering_key='saas_platform_control_base_v1'
           AND publication.offering_version_id=p_offering_version_id
           AND publication.revision=p_offering_revision
           AND publication.offering_sha256=p_offering_sha256
           AND publication.effective_at=p_offering_effective_at FOR SHARE;
    ELSIF p_evidence_kind='saas_free' THEN
        PERFORM 1 FROM platform.billing_saas_free_base_publications AS publication
         WHERE publication.offering_key='saas_free_base_v1'
           AND p_offering_key='saas_free_base_v1'
           AND publication.offering_version_id=p_offering_version_id
           AND publication.revision=p_offering_revision
           AND publication.offering_sha256=p_offering_sha256
           AND publication.effective_at=p_offering_effective_at
           AND NOT EXISTS (SELECT 1 FROM platform.billing_subscriptions AS subscription
                            WHERE subscription.tenant_id=p_tenant_id) FOR SHARE;
    ELSE
        PERFORM 1 FROM platform.billing_subscriptions AS subscription
         WHERE subscription.tenant_id=p_tenant_id
           AND subscription.offering_version_id=p_offering_version_id
           AND subscription.state IN ('trialing','active','past_due')
           AND v_subscription_sha256 ~ '^[0-9a-f]{64}$' FOR SHARE;
    END IF;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='V1 Billing evidence is not verified';
    END IF;
    INSERT INTO platform.billing_v1_base_adoption_receipts(
        operation_id,environment_id,tenant_id,assignment_id,evidence_kind,
        offering_key,offering_version_id,offering_revision,offering_sha256,
        offering_effective_at,source_sequence,source_sha256,
        subscription_evidence_sha256,plan_sha256,
        adopted_at,recorded_by
    ) VALUES (p_operation_id,p_environment_id,p_tenant_id,v_assignment_id,p_evidence_kind,
        p_offering_key,p_offering_version_id,p_offering_revision,p_offering_sha256,
        p_offering_effective_at,v_source_sequence,v_source_sha256,
        v_subscription_sha256,p_plan_sha256,
        v_now,SESSION_USER);
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_v1_base_adoption_assert_v1(
    uuid,uuid,uuid,text,text,uuid,bigint,text,timestamptz,bigint,text,text,text
) FROM PUBLIC;
