-- +goose Up
ALTER TABLE iam_system_import_sources
    ADD COLUMN delete_provider_request_id text,
    ADD COLUMN delete_response_sha256 bytea,
    ADD COLUMN absence_probe_sha256 bytea,
    ADD COLUMN verified_absent_at timestamptz,
    ADD COLUMN cleanup_state text NOT NULL DEFAULT 'waiting',
    ADD COLUMN cleanup_attempts integer NOT NULL DEFAULT 0,
    ADD COLUMN cleanup_next_at timestamptz,
    ADD COLUMN cleanup_lease_owner uuid,
    ADD COLUMN cleanup_lease_generation bigint NOT NULL DEFAULT 0,
    ADD COLUMN cleanup_lease_until timestamptz,
    ADD COLUMN cleanup_last_error_code text;

UPDATE iam_system_import_sources
SET cleanup_state=CASE WHEN state='deleted' THEN 'succeeded' ELSE 'waiting' END,
    cleanup_next_at=CASE WHEN state='deleted' THEN deleted_at
        WHEN state='claimed' THEN claimed_at ELSE delete_after END;

ALTER TABLE iam_system_import_sources
    ALTER COLUMN cleanup_next_at SET NOT NULL,
    DROP CONSTRAINT iam_system_import_sources_state_check,
    DROP CONSTRAINT iam_system_import_sources_time_check,
    ADD CONSTRAINT iam_system_import_sources_state_check CHECK (
        (state='ready' AND claimed_import_id IS NULL AND claimed_at IS NULL AND
         deleted_at IS NULL AND delete_provider_request_id IS NULL AND
         delete_response_sha256 IS NULL AND absence_probe_sha256 IS NULL AND
         verified_absent_at IS NULL) OR
        (state='claimed' AND pg_catalog.uuid_extract_version(claimed_import_id)=7 AND
         claimed_at IS NOT NULL AND deleted_at IS NULL AND
         delete_provider_request_id IS NULL AND delete_response_sha256 IS NULL AND
         absence_probe_sha256 IS NULL AND verified_absent_at IS NULL) OR
        (state='deleted' AND deleted_at IS NOT NULL AND
         pg_catalog.octet_length(delete_provider_request_id) BETWEEN 1 AND 256 AND
         pg_catalog.octet_length(delete_response_sha256)=32 AND
         pg_catalog.octet_length(absence_probe_sha256)=32 AND
         verified_absent_at IS NOT NULL)
    ),
    ADD CONSTRAINT iam_system_import_sources_time_check CHECK (
        pg_catalog.isfinite(uploaded_at) AND
        delete_after=uploaded_at+interval '24 hours' AND
        (claimed_at IS NULL OR claimed_at>=uploaded_at) AND
        (deleted_at IS NULL OR deleted_at>=uploaded_at) AND
        (verified_absent_at IS NULL OR
         (pg_catalog.isfinite(verified_absent_at) AND
          verified_absent_at=deleted_at))
    ),
    ADD CONSTRAINT iam_system_import_sources_cleanup_check CHECK (
        cleanup_attempts BETWEEN 0 AND 2147483647 AND
        cleanup_lease_generation>=0 AND pg_catalog.isfinite(cleanup_next_at) AND
        (cleanup_last_error_code IS NULL OR cleanup_last_error_code IN (
            'SOURCE_DELETE_OUTCOME_UNKNOWN','SOURCE_DELETE_UNAVAILABLE'
        )) AND (
            (state='deleted' AND cleanup_state='succeeded' AND
             cleanup_lease_owner IS NULL AND cleanup_lease_until IS NULL) OR
            (state IN ('ready','claimed') AND cleanup_state='waiting' AND
             cleanup_lease_owner IS NULL AND cleanup_lease_until IS NULL) OR
            (state IN ('ready','claimed') AND cleanup_state='running' AND
             cleanup_attempts>0 AND cleanup_lease_generation>0 AND
             pg_catalog.uuid_extract_version(cleanup_lease_owner)=7 AND
             cleanup_lease_until IS NOT NULL AND
             pg_catalog.isfinite(cleanup_lease_until))
        )
    );

-- +goose StatementBegin
CREATE FUNCTION iam_system_import_source_storage_guard_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM import source storage is immutable';
    END IF;
    IF pg_catalog.to_jsonb(NEW)-ARRAY[
           'state','claimed_import_id','claimed_at','deleted_at',
           'delete_provider_request_id','delete_response_sha256',
           'absence_probe_sha256','verified_absent_at','cleanup_state',
           'cleanup_attempts','cleanup_next_at','cleanup_lease_owner',
           'cleanup_lease_generation','cleanup_lease_until',
           'cleanup_last_error_code'
       ]::text[] IS DISTINCT FROM pg_catalog.to_jsonb(OLD)-ARRAY[
           'state','claimed_import_id','claimed_at','deleted_at',
           'delete_provider_request_id','delete_response_sha256',
           'absence_probe_sha256','verified_absent_at','cleanup_state',
           'cleanup_attempts','cleanup_next_at','cleanup_lease_owner',
           'cleanup_lease_generation','cleanup_lease_until',
           'cleanup_last_error_code'
       ]::text[] THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM import source immutable fields differ';
    END IF;
    IF (OLD.state='ready' AND OLD.cleanup_state='waiting' AND
        NEW.state='claimed' AND
        pg_catalog.uuid_extract_version(NEW.claimed_import_id)=7 AND
        NEW.claimed_at IS NOT NULL) OR
       (OLD.state=NEW.state AND NEW.state IN ('ready','claimed') AND
        OLD.cleanup_state IN ('waiting','running') AND
        NEW.cleanup_state='running' AND
        NEW.cleanup_attempts=OLD.cleanup_attempts+1 AND
        NEW.cleanup_lease_generation=OLD.cleanup_lease_generation+1 AND
        pg_catalog.uuid_extract_version(NEW.cleanup_lease_owner)=7 AND
        NEW.cleanup_lease_until>NEW.cleanup_next_at) OR
       (OLD.state=NEW.state AND NEW.state IN ('ready','claimed') AND
        OLD.cleanup_state IN ('waiting','running') AND NEW.cleanup_state='waiting' AND
        NEW.cleanup_attempts=OLD.cleanup_attempts AND
        NEW.cleanup_lease_generation=OLD.cleanup_lease_generation AND
        NEW.cleanup_lease_owner IS NULL AND NEW.cleanup_lease_until IS NULL AND
        NEW.cleanup_next_at>=OLD.cleanup_next_at) OR
       (OLD.state IN ('ready','claimed') AND NEW.state='deleted' AND
        NEW.deleted_at IS NOT NULL AND NEW.delete_provider_request_id IS NOT NULL AND
        NEW.delete_response_sha256 IS NOT NULL AND
        NEW.absence_probe_sha256 IS NOT NULL AND NEW.verified_absent_at IS NOT NULL) THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION USING ERRCODE='P6048',
        MESSAGE='IAM import source transition is immutable';
END;
$function$;

CREATE TRIGGER iam_system_import_sources_storage_guard_v1
BEFORE UPDATE ON iam_system_import_sources
FOR EACH ROW EXECUTE FUNCTION iam_system_import_source_storage_guard_v1();
REVOKE ALL ON FUNCTION iam_system_import_source_storage_guard_v1() FROM PUBLIC;

CREATE FUNCTION iam_system_import_source_cleanup_sync_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP='INSERT' THEN
        NEW.cleanup_state:='waiting';
        NEW.cleanup_attempts:=0;
        NEW.cleanup_next_at:=NEW.delete_after;
        NEW.cleanup_lease_owner:=NULL;
        NEW.cleanup_lease_generation:=0;
        NEW.cleanup_lease_until:=NULL;
        NEW.cleanup_last_error_code:=NULL;
    ELSIF NEW.state='claimed' AND OLD.state='ready' THEN
        NEW.cleanup_state:='waiting';
        NEW.cleanup_next_at:=NEW.claimed_at;
        NEW.cleanup_lease_owner:=NULL;
        NEW.cleanup_lease_until:=NULL;
        NEW.cleanup_last_error_code:=NULL;
    ELSIF NEW.state='deleted' AND OLD.state<>'deleted' THEN
        NEW.cleanup_state:='succeeded';
        NEW.cleanup_next_at:=NEW.verified_absent_at;
        NEW.cleanup_lease_owner:=NULL;
        NEW.cleanup_lease_until:=NULL;
        NEW.cleanup_last_error_code:=NULL;
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER iam_system_import_sources_zz_cleanup_sync_v1
BEFORE INSERT OR UPDATE OF state ON iam_system_import_sources
FOR EACH ROW EXECUTE FUNCTION iam_system_import_source_cleanup_sync_v1();
REVOKE ALL ON FUNCTION iam_system_import_source_cleanup_sync_v1() FROM PUBLIC;

CREATE INDEX iam_system_import_sources_cleanup_runtime_idx
ON iam_system_import_sources (cleanup_next_at,source_id)
WHERE state IN ('ready','claimed') AND cleanup_state IN ('waiting','running');
