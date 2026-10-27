-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_v1_actor_assertion_consume_v1(
    p_scope_schema name,
    p_tenant_id uuid,
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
SET search_path FROM CURRENT
AS $function$
DECLARE
	    v_authorized boolean;
    v_consumed_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v1_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM V1 actor assertion consume scope is invalid';
    END IF;
    v_authorized := iam_v1_actor_assertion_authority_v1(
        p_scope_schema, p_subject_id, p_permission_id,
        p_token_version, p_authorization_version, p_authorization_clock
    );
    IF NOT v_authorized THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM V1 actor assertion authority is denied';
    END IF;
    SELECT platform.iam_v1_actor_assertion_consume_v1(
        p_tenant_id, p_scope_schema, p_jti, p_subject_id, p_permission_id,
        p_target_domain, p_audience, p_key_id, p_permission_decision_sha256,
        p_target_binding_sha256, p_iam_group_generation, p_iam_generation,
        p_target_generation, p_token_version, p_authorization_version,
        p_authorization_clock, p_issued_at, p_expires_at
    ) INTO STRICT v_consumed_at;
    RETURN v_consumed_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_v1_actor_assertion_consume_v1(
    name,uuid,uuid,uuid,uuid,text,text,text,bytea,bytea,
    bigint,bigint,bigint,bigint,bigint,bigint,bigint,bigint
) FROM PUBLIC;
