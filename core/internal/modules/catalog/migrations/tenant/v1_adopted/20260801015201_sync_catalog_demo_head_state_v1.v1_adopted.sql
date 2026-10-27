-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_demo_sync_head_state_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_installation_id uuid;
BEGIN
    SELECT binding.installation_id INTO STRICT v_installation_id
    FROM catalog_component_bindings AS binding
    WHERE binding.component_id = NEW.component_id
      AND binding.owner_module = 'catalog'
      AND binding.owner_contract = 'catalog.demo-data.v1'
    FOR SHARE;

    INSERT INTO catalog_demo_states (
        installation_id, component_id, package_id, release_id, source_run_id,
        content_sha256, state, record_ids, last_receipt_id, revision, updated_at
    ) VALUES (
        v_installation_id, NEW.component_id, NEW.package_id, NEW.release_id,
        NEW.forms_run_id, NEW.content_sha256, 'INSTALLED', '[]'::jsonb,
        NULL, 1, NEW.activated_at
    )
    ON CONFLICT (installation_id, component_id) DO UPDATE
    SET package_id = EXCLUDED.package_id,
        release_id = EXCLUDED.release_id,
        source_run_id = EXCLUDED.source_run_id,
        content_sha256 = EXCLUDED.content_sha256,
        state = 'INSTALLED',
        record_ids = '[]'::jsonb,
        last_receipt_id = NULL,
        revision = catalog_demo_states.revision + 1,
        updated_at = EXCLUDED.updated_at
    WHERE catalog_demo_states.release_id IS DISTINCT FROM EXCLUDED.release_id
       OR catalog_demo_states.source_run_id IS DISTINCT FROM EXCLUDED.source_run_id
       OR catalog_demo_states.content_sha256 IS DISTINCT FROM EXCLUDED.content_sha256;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION catalog_demo_sync_head_state_v1() FROM PUBLIC;

CREATE TRIGGER catalog_demo_component_heads_state_sync
AFTER INSERT OR UPDATE ON catalog_demo_component_heads
FOR EACH ROW EXECUTE FUNCTION catalog_demo_sync_head_state_v1();

ALTER TABLE catalog_demo_component_heads
    ENABLE ALWAYS TRIGGER catalog_demo_component_heads_state_sync;

