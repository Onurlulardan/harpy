-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_implication_create_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_id uuid, p_source_permission_id uuid, p_implied_permission_id uuid,
    p_retirement_contract_key text, p_actor_id uuid, p_required_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (result_id uuid, result_revision bigint, authorization_clock bigint, replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_clock bigint; v_revision bigint; v_replayed boolean;
    v_count bigint; v_finished boolean; v_now timestamptz;
BEGIN
    SELECT b.authorization_clock, b.result_revision, b.replayed
    INTO v_clock, v_revision, v_replayed FROM iam_rbac_mutation_begin_v1(
        p_scope_schema, p_operation_id, 'iam.rbac.implication.create', p_request_sha256,
        p_id, p_actor_id, p_required_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) AS b;
    IF v_replayed THEN RETURN QUERY SELECT p_id, v_revision, v_clock, true; RETURN; END IF;
    v_now := pg_catalog.transaction_timestamp();
    INSERT INTO iam_permission_implications (
        id, implication_kind, source_permission_id, source_match_kind,
        implied_permission_id, implied_match_kind, retirement_contract_key,
        state, hit_count, last_hit_at, retirement_receipt_id, revision,
        created_at, updated_at, retired_at
    ) VALUES (p_id, p_retirement_contract_key, p_source_permission_id, 'exact',
        p_implied_permission_id, 'exact', p_retirement_contract_key,
        'active', 0, NULL, NULL, 1, v_now, v_now, NULL);
    v_count := iam_rbac_bump_all_users_v1(v_clock);
    v_finished := iam_rbac_mutation_finish_v1(p_operation_id,
        'iam.rbac.implication.create', p_request_sha256, v_clock,
        'iam.rbac.implication.created', p_id, 1);
    RETURN QUERY SELECT p_id, 1::bigint, v_clock, false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_rbac_implication_create_v1(
    name,uuid,bytea,uuid,uuid,uuid,text,uuid,uuid,bigint,bigint
) FROM PUBLIC;
