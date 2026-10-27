-- +goose Up
GRANT USAGE ON SCHEMA platform TO forgecrud_adopter;
GRANT USAGE ON SCHEMA platform TO forgecrud_cluster_admin;
GRANT USAGE ON SCHEMA platform TO forgecrud_runtime;

GRANT SELECT ON TABLE
    platform.jobs, platform.job_attempts, platform.outbox,
    platform.inbox, platform.idempotency_receipts, platform.dead_letters,
    platform.cutover_groups, platform.cutover_group_members,
    platform.domain_cutovers
TO forgecrud_runtime;

GRANT INSERT (
    id, tenant_id, scope_kind, scope_id, authority_group_id,
    authority_group_generation, authority_domain, cutover_generation,
    writer_origin, kind, payload_revision, payload_ref, payload_sha256,
    dedupe_key, effect_class, effect_idempotency_key, priority, run_at,
    state, attempts, max_attempts, lease_generation, retry_of_job_id
) ON TABLE platform.jobs TO forgecrud_runtime;

GRANT INSERT (
    attempt_id, job_id, scope_kind, scope_id, attempt,
    lease_generation, lease_owner_id, started_at
) ON TABLE platform.job_attempts TO forgecrud_runtime;

GRANT INSERT (
    event_id, tenant_id, scope_kind, scope_id, authority_group_id,
    authority_group_generation, authority_domain, cutover_generation,
    writer_origin, envelope_contract_version, event_type, event_version,
    aggregate_type, aggregate_id, aggregate_version, event_sequence,
    occurred_at, actor_id, correlation_id, causation_id, idempotency_key,
    business_effect_idempotency_key, payload_canonical, payload_sha256,
    payload_schema_sha256, delivery_state, delivery_attempts,
    lease_generation, next_attempt_at
) ON TABLE platform.outbox TO forgecrud_runtime;

GRANT INSERT (
    id, record_kind, consumer_name, consumer_version, tenant_id,
    scope_kind, scope_id, producer_group_id, producer_group_generation,
    producer_domain, producer_generation, writer_origin, consumer_group_id,
    consumer_group_generation, consumer_domain, consumer_generation,
    consumer_writer_origin, aggregate_type, aggregate_id, event_id,
    event_type, event_version, event_sequence, through_sequence,
    watermark_kind, causation_id, state_sha256, receipt_sha256
) ON TABLE platform.inbox TO forgecrud_runtime;

GRANT INSERT (
    receipt_id, record_kind, tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation, authority_domain,
    cutover_generation, writer_origin, request_sha256, retention_class,
    principal_id, operation_id, idempotency_key, authorization_version,
    response_status, response_headers, response_body, http_result_ref,
    response_sha256, async_job_id, effect_id, effect_kind, effect_revision,
    effect_key, provider, operation, correlation_id, causation_id,
    effect_status, effect_error_class, effect_code, provider_reference,
    effect_result_ref, effect_result_sha256, completed_at, retain_until
) ON TABLE platform.idempotency_receipts TO forgecrud_runtime;

GRANT INSERT (
    id, root_id, entry_sequence, record_kind, tenant_id, scope_kind,
    scope_id, authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin, source_kind,
    source_id, consumer_name, consumer_version, source_contract,
    source_version, source_sha256, reason_class, reason_code,
    redacted_error_sha256, payload_canonical, payload_sha256,
    payload_redacted_at, redacted_envelope, action_reason, action_ticket,
    action_actor_id, action_causation_id, full_payload_expires_at,
    metadata_expires_at, recorded_at
) ON TABLE platform.dead_letters TO forgecrud_runtime;

GRANT UPDATE (
    state, attempts, lease_generation, lease_owner_id,
    lease_heartbeat_at, lease_until, cancel_requested_at,
    effect_started_at, run_at, last_error_class, last_error_code,
    last_error_sha256, retain_until, updated_at, completed_at
) ON TABLE platform.jobs TO forgecrud_runtime;

GRANT UPDATE (
    finished_at, outcome, error_class, error_code,
    redacted_error_sha256, effect_started_at, effect_id,
    effect_request_sha256, effect_receipt_id
) ON TABLE platform.job_attempts TO forgecrud_runtime;

GRANT UPDATE (
    delivery_state, delivery_attempts, lease_owner_id, lease_generation,
    lease_heartbeat_at, lease_until, next_attempt_at, confirm_sha256,
    last_error_code, last_error_sha256, updated_at
) ON TABLE platform.outbox TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION
    platform.platform_async_assert_current_authority_v1(
        text,uuid,text,bigint,text,bigint
    )
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_park_stale_work_v1(
        uuid,uuid,bigint,uuid,bytea,text,jsonb,text
    )
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_release_job_retention_v1(
        uuid,uuid,text,bigint,text,bigint,bigint,text,timestamptz
    )
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_validate_job_attempt_v1(
        platform.job_attempts,platform.job_attempts,name,name
    )
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_validate_job_transition_v1(
        platform.jobs,platform.jobs,name,name
    )
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_begin_v1_adoption_v1(
        uuid,text,uuid,text,text,bigint,bigint,bigint,bigint,text,text,
        smallint,smallint,smallint,bigint,timestamptz
    )
TO forgecrud_adopter;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_record_v1_inventory_item_v1(
        uuid,text,text,text,text,text,integer,text,text,text,text,integer,
        text,text,timestamptz,integer,integer,text,text,timestamptz,text
    )
TO forgecrud_adopter;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_seal_v1_inventory_v1(uuid,text)
TO forgecrud_adopter;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_adopt_v1_inventory_item_v1(
        uuid,text,text,text,text,uuid,timestamptz,text
    )
TO forgecrud_adopter;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_cancel_v1_inventory_item_v1(
        uuid,text,text,text,text,text,text,text,timestamptz,text
    )
TO forgecrud_adopter;

GRANT EXECUTE ON FUNCTION
    platform.platform_async_record_v1_restart_fence_v1(
        uuid,text,text,text,text,text,text,timestamptz,timestamptz,text
    )
TO forgecrud_cluster_admin;

GRANT EXECUTE ON FUNCTION
    platform.platform_async_record_v1_zero_inventory_v1(
        uuid,text,bigint,bigint,bigint,bigint,timestamptz,text
    )
TO forgecrud_cluster_admin;

GRANT EXECUTE ON FUNCTION
    platform.platform_async_accept_v1_adoption_v1(
        uuid,text,text,timestamptz,text
    )
TO forgecrud_cluster_admin;
