-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_apply_v1_provider_account_v1(
    p_operation_id uuid, p_run_id uuid, p_record jsonb
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_id uuid;
    v_sha bytea;
BEGIN
    v_id := (p_record #>> '{}')::uuid;
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
      AND run.scope_kind = 'global' AND run.state = 'planned' AND run.eligible;
    INSERT INTO platform.billing_provider_accounts (
        id, authority_id, provider_key, provider_account_key, lifecycle,
        supports_checkout, supports_webhooks, supports_proration,
        supports_refunds, created_at, disabled_at, revision
    ) VALUES (
        v_id, v_run.authority_id, 'manual', 'default', 'active',
        false, false, false, false, v_run.snapshot_at, NULL, 1
    );
    v_sha := platform.billing_v1_adoption_catalog_fingerprint_v1(
        'provider_account', v_id
    );
    PERFORM platform.billing_record_v1_adoption_mapping_v1(
        p_operation_id, p_run_id, 'provider_accounts', 'manual/default',
        'provider_account', v_id, v_sha, v_run.snapshot_at
    );
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_apply_v1_provider_account_v1(
    uuid, uuid, jsonb
) FROM PUBLIC;

CREATE FUNCTION platform.billing_apply_v1_offering_v1(
    p_operation_id uuid, p_run_id uuid, p_record jsonb
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_price jsonb;
    v_grant jsonb;
    v_definition platform.billing_entitlement_definitions%ROWTYPE;
    v_registry bigint;
    v_metadata jsonb;
    v_sha bytea;
    v_retired_at timestamptz;
BEGIN
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
      AND run.scope_kind = 'global' AND run.state = 'planned' AND run.eligible;
    v_retired_at := NULLIF(p_record->>'retired_at', '')::timestamptz;
    v_metadata := pg_catalog.jsonb_build_object(
        'name', p_record->>'name', 'description', p_record->>'description',
        'sort_order', (p_record->>'sort_order')::integer,
        'highlight_label', p_record->>'highlight_label',
        'color', p_record->>'color',
        'billing_interval', p_record->>'billing_interval'
    );
    INSERT INTO platform.billing_offerings (
        id, authority_id, offering_key, lifecycle, created_at, retired_at, revision
    ) VALUES (
        (p_record->>'id')::uuid, v_run.authority_id, p_record->>'key',
		'active', (p_record->>'created_at')::timestamptz, NULL, 1
    );
    INSERT INTO platform.billing_offering_versions (
		id, authority_id, offering_id, version, is_base, is_public, available,
        trial_days, grace_days, expected_price_count, expected_grant_count,
        effective_start, effective_end, presentation_metadata,
        presentation_sha256, publication_state, publication_operation_id, published_at
    ) VALUES (
        (p_record->>'version_id')::uuid, v_run.authority_id,
        (p_record->>'id')::uuid, 1, (p_record->>'is_base')::boolean,
		(p_record->>'is_public')::boolean, (p_record->>'available')::boolean,
		(p_record->>'trial_days')::integer, 0,
        pg_catalog.jsonb_array_length(p_record->'prices'),
        pg_catalog.jsonb_array_length(p_record->'grants'),
        (p_record->>'created_at')::timestamptz,
		v_retired_at,
        v_metadata, pg_catalog.sha256(pg_catalog.convert_to(v_metadata::text, 'UTF8')),
        'published', (p_record->>'publication_operation_id')::uuid,
        (p_record->>'updated_at')::timestamptz
    );
    FOR v_price IN SELECT value FROM pg_catalog.jsonb_array_elements(p_record->'prices') LOOP
        SELECT exponent.registry_version INTO STRICT v_registry
        FROM platform.billing_currency_exponents AS exponent
        WHERE exponent.authority_id = v_run.authority_id
          AND exponent.code = (v_price->>'currency')::character(3)
          AND exponent.financial_money_eligible
        ORDER BY exponent.registry_version DESC LIMIT 1;
        INSERT INTO platform.billing_price_versions (
            id, authority_id, offering_version_id, cadence, minor_units,
            currency, currency_registry_version, currency_eligible,
            effective_start, effective_end, money_policy_version, published_at
        ) VALUES (
            (v_price->>'id')::uuid, v_run.authority_id,
            (p_record->>'version_id')::uuid, v_price->>'cadence',
            (v_price->>'minor_units')::bigint, (v_price->>'currency')::character(3),
            v_registry, true, (p_record->>'created_at')::timestamptz,
			v_retired_at,
            'fc_money_v1', (p_record->>'updated_at')::timestamptz
        );
    END LOOP;
    FOR v_grant IN SELECT value FROM pg_catalog.jsonb_array_elements(p_record->'grants') LOOP
        SELECT definition.* INTO STRICT v_definition
        FROM platform.billing_entitlement_definitions AS definition
        WHERE definition.authority_id = v_run.authority_id
          AND definition.entitlement_key = v_grant->>'entitlement_key';
        INSERT INTO platform.billing_offering_entitlements (
            authority_id, offering_version_id, entitlement_definition_id,
            value_type, definition_allows_unlimited, boolean_value,
            integer_value, unlimited, created_at
        ) VALUES (
            v_run.authority_id, (p_record->>'version_id')::uuid, v_definition.id,
            v_grant->>'value_type', v_definition.unlimited_allowed,
            NULLIF(v_grant->>'boolean_value', '')::boolean,
            NULLIF(v_grant->>'integer_value', '')::bigint,
            (v_grant->>'unlimited')::boolean, (p_record->>'updated_at')::timestamptz
        );
    END LOOP;
    IF v_retired_at IS NOT NULL THEN
        UPDATE platform.billing_offerings
        SET lifecycle = 'retired', retired_at = v_retired_at, revision = 2
        WHERE authority_id = v_run.authority_id
          AND id = (p_record->>'id')::uuid
          AND lifecycle = 'active' AND revision = 1;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'billing adopted offering retirement failed';
        END IF;
    END IF;
    v_sha := platform.billing_v1_adoption_catalog_fingerprint_v1(
        'offering', (p_record->>'id')::uuid
    );
    PERFORM platform.billing_record_v1_adoption_mapping_v1(
        p_operation_id, p_run_id, 'plans', p_record->>'source_id',
        'offering', (p_record->>'id')::uuid, v_sha, v_run.snapshot_at
    );
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_apply_v1_offering_v1(
    uuid, uuid, jsonb
) FROM PUBLIC;
-- +goose StatementEnd
