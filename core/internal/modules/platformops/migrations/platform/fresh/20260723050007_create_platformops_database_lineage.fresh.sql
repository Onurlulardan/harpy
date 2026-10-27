-- +goose Up
CREATE TABLE platform.platformops_database_lineages (
    environment_id uuid PRIMARY KEY,
    lineage_id uuid NOT NULL UNIQUE,
    current_database_system_id uuid NOT NULL UNIQUE,
    database_system_identifier text NOT NULL,
    generation bigint NOT NULL CHECK (generation > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at)
);

CREATE TABLE platform.platformops_database_lineage_events (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    previous_database_system_id uuid,
    database_system_id uuid NOT NULL UNIQUE,
    generation bigint NOT NULL CHECK (generation > 0),
    event_kind text NOT NULL CHECK (event_kind IN ('initialized','restored')),
    occurred_at timestamptz NOT NULL,
    CONSTRAINT platformops_database_lineage_events_lineage_fk
        FOREIGN KEY (environment_id)
        REFERENCES platform.platformops_database_lineages(environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TRIGGER platformops_database_lineage_events_append_only
BEFORE UPDATE OR DELETE ON platform.platformops_database_lineage_events
FOR EACH ROW EXECUTE FUNCTION platform.platformops_reject_append_mutation_v1();

-- +goose StatementBegin
CREATE FUNCTION platform.platformops_initialize_database_lineage_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_system_identifier text;
    v_database_system_id uuid;
BEGIN
    IF NEW.deployment_mode <> 'onpremise' THEN
        RETURN NEW;
    END IF;
    SELECT identity.database_system_identifier
      INTO v_system_identifier
      FROM platform.environment_identity AS identity
     WHERE identity.environment_id = NEW.environment_id;
    IF v_system_identifier IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'database lineage environment identity is missing';
    END IF;
    v_database_system_id := pg_catalog.gen_random_uuid();
    INSERT INTO platform.platformops_database_lineages(
        environment_id,lineage_id,current_database_system_id,
        database_system_identifier,generation,created_at,updated_at
    ) VALUES (
        NEW.environment_id,NEW.environment_id,v_database_system_id,
        v_system_identifier,1,NEW.created_at,NEW.created_at
    );
    INSERT INTO platform.platformops_database_lineage_events(
        operation_id,environment_id,previous_database_system_id,
        database_system_id,generation,event_kind,occurred_at
    ) VALUES (
        NEW.operation_id,NEW.environment_id,NULL,v_database_system_id,
        1,'initialized',NEW.created_at
    );
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.platformops_current_environment_identity_v1(
    p_expected_environment_id uuid
)
RETURNS TABLE(
    environment_id uuid,deployment_mode text,primary_binding_id uuid,
    tenant_set_sha256 text,owner_binding_id uuid,
    owner_binding_generation bigint,database_lineage_id uuid,
    database_system_id uuid
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT identity.environment_id,identity.deployment_mode,
           target.tenant_id,pg_catalog.encode(target.tenant_set_sha256,'hex'),
           owner.owner_binding_id,owner.owner_policy_revision,
           lineage.lineage_id,lineage.current_database_system_id
      FROM platform.platformops_installation_recovery_identities AS identity
      JOIN platform.tenancy_get_primary_tenant_target_v2() AS target
        ON target.environment_id = identity.environment_id
       AND target.tenant_id = identity.tenant_id
       AND target.deployment_mode = identity.deployment_mode
       AND target.purpose = 'onpremise_primary'
      JOIN platform.iam_platform_current_owner_projection_v1(
          identity.environment_id
      ) AS owner ON owner.environment_id = identity.environment_id
      JOIN platform.platformops_database_lineages AS lineage
        ON lineage.environment_id = identity.environment_id
     WHERE identity.environment_id = p_expected_environment_id
       AND identity.deployment_mode = 'onpremise'
       AND identity.identity_state = 'active';
END;
$function$;

CREATE FUNCTION platform.platformops_commit_database_lineage_restore_v1(
    p_operation_id uuid,p_expected_database_system_id uuid,
    p_new_database_system_id uuid,p_occurred_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_restore platform.platformops_restore_operations%ROWTYPE;
    v_generation bigint;
BEGIN
    SELECT * INTO v_restore
      FROM platform.platformops_restore_operations AS restore
     WHERE restore.operation_id = p_operation_id
       AND restore.state = 'completed'
       AND restore.new_database_system_id = p_new_database_system_id
     FOR SHARE;
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    UPDATE platform.platformops_database_lineages
       SET current_database_system_id = p_new_database_system_id,
           generation = generation + 1,updated_at = p_occurred_at
     WHERE environment_id = v_restore.environment_id
       AND current_database_system_id = p_expected_database_system_id
       AND p_new_database_system_id <> p_expected_database_system_id
    RETURNING generation INTO v_generation;
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    INSERT INTO platform.platformops_database_lineage_events(
        operation_id,environment_id,previous_database_system_id,
        database_system_id,generation,event_kind,occurred_at
    ) VALUES (
        p_operation_id,v_restore.environment_id,p_expected_database_system_id,
        p_new_database_system_id,v_generation,'restored',p_occurred_at
    );
    RETURN true;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER platformops_installation_database_lineage_initialize
AFTER INSERT ON platform.platformops_installation_recovery_identities
FOR EACH ROW EXECUTE FUNCTION platform.platformops_initialize_database_lineage_v1();

REVOKE ALL ON FUNCTION platform.platformops_current_environment_identity_v1(uuid)
FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_commit_database_lineage_restore_v1(
    uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_current_environment_identity_v1(uuid)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_current_environment_identity_v1(uuid)
TO forgecrud_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_commit_database_lineage_restore_v1(
    uuid,uuid,uuid,timestamptz
) TO forgecrud_restore;
