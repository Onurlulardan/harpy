-- +goose Up
ALTER TABLE catalog_installations
    ADD CONSTRAINT catalog_installations_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE catalog_installations
    VALIDATE CONSTRAINT catalog_installations_package_id_nonzero;

ALTER TABLE catalog_dependency_resolutions
    ADD CONSTRAINT catalog_dependencies_package_id_nonzero CHECK (
        dependency_package_id IS NOT NULL AND
        dependency_package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE catalog_dependency_resolutions
    VALIDATE CONSTRAINT catalog_dependencies_package_id_nonzero;

ALTER TABLE catalog_operation_runs
    ADD CONSTRAINT catalog_operation_runs_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE catalog_operation_runs
    VALIDATE CONSTRAINT catalog_operation_runs_package_id_nonzero;

ALTER TABLE catalog_command_receipts
    ADD CONSTRAINT catalog_command_receipts_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE catalog_command_receipts
    VALIDATE CONSTRAINT catalog_command_receipts_package_id_nonzero;

ALTER TABLE catalog_menu_items
    ADD CONSTRAINT catalog_menu_items_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE catalog_menu_items
    VALIDATE CONSTRAINT catalog_menu_items_package_id_nonzero;

ALTER TABLE catalog_tombstones
    ADD CONSTRAINT catalog_tombstones_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE catalog_tombstones
    VALIDATE CONSTRAINT catalog_tombstones_package_id_nonzero;

ALTER TABLE catalog_demo_payloads
    ADD CONSTRAINT catalog_demo_payloads_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE catalog_demo_payloads
    VALIDATE CONSTRAINT catalog_demo_payloads_package_id_nonzero;

ALTER TABLE catalog_demo_provisions
    ADD CONSTRAINT catalog_demo_provisions_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE catalog_demo_provisions
    VALIDATE CONSTRAINT catalog_demo_provisions_package_id_nonzero;

ALTER TABLE catalog_demo_handoffs
    ADD CONSTRAINT catalog_demo_handoffs_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE catalog_demo_handoffs
    VALIDATE CONSTRAINT catalog_demo_handoffs_package_id_nonzero;

ALTER TABLE catalog_demo_component_heads
    ADD CONSTRAINT catalog_demo_heads_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE catalog_demo_component_heads
    VALIDATE CONSTRAINT catalog_demo_heads_package_id_nonzero;

ALTER TABLE catalog_purge_authorities_v2
    ADD CONSTRAINT catalog_purge_authority_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE catalog_purge_authorities_v2
    VALIDATE CONSTRAINT catalog_purge_authority_package_id_nonzero;

ALTER TABLE catalog_purge_approval_receipts_v2
    ADD CONSTRAINT catalog_purge_approval_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE catalog_purge_approval_receipts_v2
    VALIDATE CONSTRAINT catalog_purge_approval_package_id_nonzero;
