-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_v1_actor_assertion_consume_v1(
    p_tenant_id uuid,
    p_scope_schema name,
    p_jti uuid,
    p_subject_id uuid,
    p_permission_id uuid,
    p_target_domain text,
    p_audience text,
    p_key_id text,
    p_permission_decision_sha256 bytea,
    p_target_binding_sha256 bytea,
    p_iam_group_generation bigint,
    p_iam_generation bigint,
    p_target_generation bigint,
    p_token_version bigint,
    p_authorization_version bigint,
    p_authorization_clock bigint,
    p_issued_at bigint,
    p_expires_at bigint
)
RETURNS timestamptz
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_iam_group_generation bigint;
    v_iam_generation bigint;
    v_target_generation bigint;
    v_retirement platform.contract_retirements%ROWTYPE;
    v_schema text;
    v_issued_at timestamptz;
    v_expires_at timestamptz;
    v_now timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v1_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_tenant_id IS NULL OR p_scope_schema IS NULL OR
       p_jti IS NULL OR p_subject_id IS NULL OR
       p_permission_id IS NULL OR p_target_domain IS NULL OR p_audience IS NULL OR
       p_key_id IS NULL OR p_permission_decision_sha256 IS NULL OR
       p_target_binding_sha256 IS NULL OR p_iam_group_generation < 1 OR
       p_iam_generation < 1 OR p_target_generation < 1 OR p_token_version < 1 OR
       p_authorization_version < 1 OR p_authorization_clock < 0 OR
       p_issued_at < 1 OR p_expires_at - p_issued_at <> 30 OR
       p_audience IS DISTINCT FROM
           'urn:forgecrud:v1-domain:' || p_target_domain OR
       p_key_id !~ '^[a-z][a-z0-9_-]{0,62}$' OR
       pg_catalog.octet_length(p_permission_decision_sha256) <> 32 OR
       pg_catalog.octet_length(p_target_binding_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 actor assertion consume input is invalid';
    END IF;

    SELECT authority.iam_group_generation, authority.iam_generation,
           authority.target_generation
    INTO STRICT v_iam_group_generation, v_iam_generation, v_target_generation
    FROM platform.iam_v1_actor_bridge_authority_v1(
        p_tenant_id, p_target_domain
    ) AS authority;
    IF v_iam_group_generation IS DISTINCT FROM p_iam_group_generation OR
       v_iam_generation IS DISTINCT FROM p_iam_generation OR
       v_target_generation IS DISTINCT FROM p_target_generation THEN
        RAISE EXCEPTION USING ERRCODE = 'P5602',
            MESSAGE = 'IAM V1 actor assertion generation is stale';
    END IF;

    SELECT retirement.* INTO v_retirement
    FROM platform.contract_retirements AS retirement
    WHERE retirement.contract_key = 'iam.v1_actor_assertion.v1'
    FOR SHARE;
    IF NOT FOUND OR v_retirement.owner_module IS DISTINCT FROM 'iam' OR
       v_retirement.domain IS DISTINCT FROM 'iam' OR
       v_retirement.state IS DISTINCT FROM 'active' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5603',
            MESSAGE = 'IAM V1 actor bridge contract is unavailable';
    END IF;

    v_issued_at := pg_catalog.to_timestamp(p_issued_at);
    v_expires_at := pg_catalog.to_timestamp(p_expires_at);
    v_now := pg_catalog.clock_timestamp();
    IF v_expires_at IS DISTINCT FROM v_issued_at + INTERVAL '30 seconds' OR
       v_now < v_issued_at - INTERVAL '2 seconds' OR v_now >= v_expires_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P5605',
            MESSAGE = 'IAM V1 actor assertion lifetime is invalid';
    END IF;

    SELECT tenant.schema_name::text INTO v_schema
    FROM platform.tenants AS tenant
    WHERE tenant.id = p_tenant_id AND tenant.deleted_at IS NULL
    FOR SHARE;
    IF NOT FOUND OR v_schema IS NULL OR
       v_schema IS DISTINCT FROM p_scope_schema::text OR
       v_schema !~ '^t_[a-z2-7]{26}$|^[a-z_][a-z0-9_]{0,62}$' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5601',
            MESSAGE = 'IAM V1 actor assertion tenant is unavailable';
    END IF;

    BEGIN
        INSERT INTO platform.iam_v1_actor_assertion_receipts (
            tenant_id, jti, subject_id, permission_id, target_domain, audience,
            key_id, permission_decision_sha256, target_binding_sha256,
            iam_group_generation, iam_generation, target_generation,
            token_version, authorization_version, authorization_clock,
            issued_at, expires_at, consumed_at
        ) VALUES (
            p_tenant_id, p_jti, p_subject_id, p_permission_id, p_target_domain,
            p_audience, p_key_id, p_permission_decision_sha256,
            p_target_binding_sha256, p_iam_group_generation, p_iam_generation,
            p_target_generation, p_token_version, p_authorization_version,
            p_authorization_clock, v_issued_at, v_expires_at, v_now
        );
    EXCEPTION WHEN unique_violation THEN
        RAISE EXCEPTION USING ERRCODE = 'P5604',
            MESSAGE = 'IAM V1 actor assertion replay is denied';
    END;
    RETURN v_now;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_v1_actor_assertion_consume_v1(
    uuid,name,uuid,uuid,uuid,text,text,text,bytea,bytea,
    bigint,bigint,bigint,bigint,bigint,bigint,bigint,bigint
) FROM PUBLIC;
