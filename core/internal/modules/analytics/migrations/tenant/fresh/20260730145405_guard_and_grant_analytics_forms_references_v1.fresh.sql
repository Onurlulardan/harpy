-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_forms_reference_guard_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    IF TG_OP='DELETE' OR
       NEW.reference_id<>OLD.reference_id OR
       NEW.register_operation_id<>OLD.register_operation_id OR
       NEW.source_kind<>OLD.source_kind OR NEW.source_id<>OLD.source_id OR
       NEW.source_version_id<>OLD.source_version_id OR
       NEW.component_id IS DISTINCT FROM OLD.component_id OR
       NEW.target_definition_id<>OLD.target_definition_id OR
       NEW.target_version_id<>OLD.target_version_id OR
       NEW.usage<>OLD.usage OR NEW.claimed_by<>OLD.claimed_by OR
       NEW.claimed_at<>OLD.claimed_at OR
       OLD.target_reference_epoch IS NOT NULL AND
       (NEW.target_reference_epoch<>OLD.target_reference_epoch OR
        NEW.register_result_sha256<>OLD.register_result_sha256 OR
        NEW.registered_at<>OLD.registered_at) OR
       OLD.release_operation_id IS NOT NULL AND
       (NEW.release_operation_id<>OLD.release_operation_id OR
        NEW.release_claimed_at<>OLD.release_claimed_at) OR
       OLD.released_at IS NOT NULL AND
       (NEW.release_result_sha256<>OLD.release_result_sha256 OR
        NEW.released_at<>OLD.released_at) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Forms reference evidence is immutable';
    END IF;
    RETURN NEW;
END;
$function$;

-- +goose StatementEnd

CREATE TRIGGER analytics_forms_reference_evidence_guard
BEFORE UPDATE OR DELETE ON analytics_forms_reference_bindings
FOR EACH ROW EXECUTE FUNCTION analytics_forms_reference_guard_v1();
ALTER TABLE analytics_forms_reference_bindings
    ENABLE ALWAYS TRIGGER analytics_forms_reference_evidence_guard;

REVOKE ALL ON FUNCTION analytics_forms_reference_guard_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_forms_reference_claim_v1(
    name,uuid,uuid,text,uuid,uuid,uuid,uuid,uuid,uuid,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_forms_reference_register_ack_v1(
    name,uuid,uuid,bigint,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_forms_reference_active_v1(
    name,uuid,uuid,uuid
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_forms_reference_release_claim_v1(
    name,uuid,uuid,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_forms_reference_release_ack_v1(
    name,uuid,uuid,bigint,bytea,timestamptz
) TO forgecrud_runtime;
