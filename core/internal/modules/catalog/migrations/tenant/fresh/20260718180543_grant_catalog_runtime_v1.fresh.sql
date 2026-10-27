-- +goose Up
GRANT SELECT ON TABLE catalog_installations, catalog_component_bindings,
    catalog_dependency_resolutions, catalog_operation_runs, catalog_operation_steps,
    catalog_command_receipts, catalog_purge_blocker_receipts,
    catalog_menu_definitions, catalog_menu_versions, catalog_menu_items,
    catalog_menu_overrides, catalog_menu_override_receipts, catalog_adoption_runs,
    catalog_legacy_route_hits, catalog_legacy_consumer_acks, catalog_tombstones
TO forgecrud_runtime;

GRANT INSERT (installation_id, package_id, active_version, state, reference_epoch,
    row_version, updated_at) ON catalog_installations TO forgecrud_runtime;
GRANT UPDATE (active_version, state, reference_epoch, row_version, updated_at)
    ON catalog_installations TO forgecrud_runtime;
GRANT INSERT (installation_id, component_id, owner_module, owner_generation,
    owner_contract, content_sha256, state) ON catalog_component_bindings TO forgecrud_runtime;
GRANT UPDATE (owner_module, owner_generation, owner_contract, content_sha256, state)
    ON catalog_component_bindings TO forgecrud_runtime;
GRANT INSERT (installation_id, dependency_package_id, exact_version, required, plan_sha256)
    ON catalog_dependency_resolutions TO forgecrud_runtime;
GRANT UPDATE (exact_version, required, plan_sha256)
    ON catalog_dependency_resolutions TO forgecrud_runtime;
GRANT INSERT (operation_id, installation_id, package_id, kind, state, plan, plan_sha256,
    retry_of_operation_id, actor_id, idempotency_key, request_sha256, started_at,
    destructive_started_at, failure_code, revision) ON catalog_operation_runs TO forgecrud_runtime;
GRANT UPDATE (state, destructive_started_at, failure_code, revision)
    ON catalog_operation_runs TO forgecrud_runtime;
GRANT INSERT (receipt_id, operation_id, component_id, owner_module, owner_generation,
    owner_contract, content_sha256, action, ready, completed_at, receipt_sha256)
    ON catalog_operation_steps TO forgecrud_runtime;
GRANT INSERT (receipt_id, operation_id, package_id, component_id, owner_module,
    owner_generation, owner_contract, content_sha256, action, idempotency_key, completed_at)
    ON catalog_command_receipts TO forgecrud_runtime;
GRANT INSERT (receipt_id, operation_id, component_id, reference_epoch, catalog_version,
    zero_digest, completed_at) ON catalog_purge_blocker_receipts TO forgecrud_runtime;

GRANT INSERT (definition_id, revision, active_version_id, updated_at)
    ON catalog_menu_definitions TO forgecrud_runtime;
GRANT UPDATE (active_version_id, revision, updated_at)
    ON catalog_menu_definitions TO forgecrud_runtime;
GRANT INSERT (version_id, definition_id, version_number, content_sha256, created_at)
    ON catalog_menu_versions TO forgecrud_runtime;
GRANT INSERT (version_id, item_id, parent_item_id, package_id, item_kind, target_id,
    external_url, label, sort_order) ON catalog_menu_items TO forgecrud_runtime;
GRANT INSERT (item_id, label, sort_order, visible, revision)
    ON catalog_menu_overrides TO forgecrud_runtime;
GRANT UPDATE (label, sort_order, visible, revision)
    ON catalog_menu_overrides TO forgecrud_runtime;
GRANT INSERT (receipt_id, actor_id, idempotency_key, request_sha256, item_id,
    result_revision, result_label, result_sort_order, result_visible, completed_at)
    ON catalog_menu_override_receipts TO forgecrud_runtime;

GRANT INSERT (run_id, source_manifest_sha256, transform_sha256, source_count, target_count,
    source_root_sha256, target_root_sha256, state, blocker_code, cutover_generation,
    created_at, verified_at) ON catalog_adoption_runs TO forgecrud_runtime;
GRANT INSERT (route_key, hit_count, last_hit_at, consumer_ack_count,
    required_consumer_count, remove_after) ON catalog_legacy_route_hits TO forgecrud_runtime;
GRANT UPDATE (hit_count, last_hit_at, consumer_ack_count, remove_after, retired_at)
    ON catalog_legacy_route_hits TO forgecrud_runtime;
GRANT INSERT (route_key, consumer_id, evidence_sha256, acknowledged_at)
    ON catalog_legacy_consumer_acks TO forgecrud_runtime;
GRANT INSERT (package_id, slug, package_sha256, reference_epoch, owner_receipt_root, purged_at)
    ON catalog_tombstones TO forgecrud_runtime;

-- +goose StatementBegin
CREATE FUNCTION catalog_delete_dependency_resolutions_v1(p_installation_id uuid)
RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_deleted bigint;
BEGIN
    DELETE FROM catalog_dependency_resolutions
    WHERE installation_id = p_installation_id;
    GET DIAGNOSTICS v_deleted = ROW_COUNT;
    RETURN v_deleted;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION catalog_delete_dependency_resolutions_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION catalog_delete_dependency_resolutions_v1(uuid)
    TO forgecrud_runtime;
