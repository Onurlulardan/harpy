-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.observe_finalizer_authority_v1(
    p_operation_id uuid, p_run_id uuid
)
RETURNS TABLE (
    finalization_state_sha256 text,
    finalization_procedure_oid oid,
    finalization_procedure_signature text,
    database_name text,
    database_origin text,
    system_identifier text,
    database_oid oid,
    operation_state text,
    completed_eligible boolean,
    abort_eligible boolean,
    completed_tenant_catalog_sha256 text,
    completed_tenant_requirement_count integer,
    finalizer_valid_until timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.cutover_operations%ROWTYPE;
    v_finalizer_role pg_catalog.pg_roles%ROWTYPE;
    v_state_sha256 text;
    v_database_origin text;
    v_system_identifier text;
    v_database_oid oid;
    v_catalog_sha256 text;
    v_requirement_count integer;
    v_authority_receipt_sha256 text;
BEGIN
    IF SESSION_USER <> 'harpy_cluster_admin' OR p_operation_id IS NULL OR p_run_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='cutover finalizer authority observer denied';
    END IF;
    SELECT inventory.receipt_sha256 INTO STRICT v_authority_receipt_sha256
    FROM cutover_authority.authority_adoption_inventory AS inventory
    ORDER BY inventory.catalog_generation DESC LIMIT 1;
    PERFORM cutover_authority.assert_adopted_catalog_v1(v_authority_receipt_sha256);
    SELECT role.* INTO STRICT v_finalizer_role FROM pg_catalog.pg_roles AS role
    WHERE role.rolname='harpy_cutover_finalizer' AND role.rolcanlogin AND NOT role.rolinherit
      AND NOT role.rolsuper AND NOT role.rolcreatedb AND NOT role.rolcreaterole
      AND NOT role.rolreplication AND NOT role.rolbypassrls AND role.rolconnlimit=1
      AND role.rolvaliduntil>pg_catalog.clock_timestamp()
      AND role.rolvaliduntil<=pg_catalog.clock_timestamp()+interval '120 seconds'
      AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
          WHERE membership.roleid=role.oid OR membership.member=role.oid);
    v_state_sha256:=cutover_authority.finalization_state_sha256_v1(p_operation_id,p_run_id);
    SELECT operation.* INTO STRICT v_operation FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id;
    SELECT release_row.database_origin,control.system_identifier::text,database_row.oid,
           catalog.catalog_sha256,catalog.requirement_count
    INTO STRICT v_database_origin,v_system_identifier,v_database_oid,
         v_catalog_sha256,v_requirement_count
    FROM platform.release_generations AS release_row
    CROSS JOIN pg_catalog.pg_control_system() AS control
    CROSS JOIN pg_catalog.pg_database AS database_row
    LEFT JOIN platform.tenant_migration_requirement_catalogs AS catalog
      ON catalog.migration_run_id=release_row.migration_run_id
     AND catalog.release_id=release_row.release_id
     AND catalog.environment_id=release_row.environment_id
     AND catalog.database_origin=release_row.database_origin
    WHERE release_row.release_id=v_operation.release_id
      AND release_row.environment_id=v_operation.environment_id
      AND database_row.datname=pg_catalog.current_database();
    RETURN QUERY SELECT v_state_sha256,
        'cutover_authority.finalize_operation_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)'::regprocedure::oid,
        'cutover_authority.finalize_operation_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)'::text,
        pg_catalog.current_database()::text,v_database_origin,v_system_identifier,v_database_oid,
        v_operation.state,
        v_operation.state='quiescing' AND v_catalog_sha256 IS NOT NULL AND v_requirement_count IS NOT NULL,
        v_operation.state IN ('planned','copying','shadow','quiescing','blocked'),
        v_catalog_sha256,v_requirement_count,v_finalizer_role.rolvaliduntil;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.observe_finalizer_authority_v1(uuid,uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION cutover_authority.observe_finalizer_authority_v1(uuid,uuid) TO harpy_cluster_admin;
