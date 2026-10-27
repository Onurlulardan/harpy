SELECT
    session_principal.rolname,
    session_principal.rolsuper,
    session_principal.rolinherit,
    session_principal.rolcreaterole,
    session_principal.rolcreatedb,
    session_principal.rolcanlogin,
    session_principal.rolreplication,
    session_principal.rolbypassrls,
    session_principal.rolconnlimit,
    current_principal.rolname,
    current_principal.rolsuper,
    current_principal.rolinherit,
    current_principal.rolcreaterole,
    current_principal.rolcreatedb,
    current_principal.rolcanlogin,
    current_principal.rolreplication,
    current_principal.rolbypassrls,
    current_principal.rolconnlimit,
    COALESCE(membership.admin_option, false),
    COALESCE(membership.inherit_option, false),
    COALESCE(membership.set_option, false),
    (SELECT pg_catalog.count(*) FROM pg_catalog.pg_auth_members AS all_memberships
        WHERE all_memberships.member = session_principal.oid),
    EXISTS (
        SELECT 1
        FROM pg_catalog.pg_auth_members AS capability_membership
        JOIN pg_catalog.pg_roles AS capability_role
          ON capability_role.oid = capability_membership.roleid
        WHERE capability_membership.member = session_principal.oid
          AND capability_role.rolname = 'harpy_forms_record_reader'
          AND NOT capability_membership.admin_option
          AND NOT capability_membership.inherit_option
          AND capability_membership.set_option
          AND NOT capability_role.rolcanlogin AND NOT capability_role.rolinherit
          AND NOT capability_role.rolsuper AND NOT capability_role.rolcreatedb
          AND NOT capability_role.rolcreaterole AND NOT capability_role.rolreplication
          AND NOT capability_role.rolbypassrls AND capability_role.rolconnlimit = -1
    ),
    EXISTS (
        SELECT 1
        FROM pg_catalog.pg_auth_members AS capability_membership
        JOIN pg_catalog.pg_roles AS capability_role
          ON capability_role.oid = capability_membership.roleid
        WHERE capability_membership.member = session_principal.oid
          AND capability_role.rolname = 'harpy_forms_record_writer'
          AND NOT capability_membership.admin_option
          AND NOT capability_membership.inherit_option
          AND capability_membership.set_option
          AND NOT capability_role.rolcanlogin AND NOT capability_role.rolinherit
          AND NOT capability_role.rolsuper AND NOT capability_role.rolcreatedb
          AND NOT capability_role.rolcreaterole AND NOT capability_role.rolreplication
          AND NOT capability_role.rolbypassrls AND capability_role.rolconnlimit = -1
    ),
    pg_catalog.has_database_privilege(session_principal.rolname, pg_catalog.current_database(), 'TEMP'),
    pg_catalog.has_database_privilege(current_principal.rolname, pg_catalog.current_database(), 'TEMP'),
    pg_catalog.has_database_privilege(session_principal.rolname, pg_catalog.current_database(), 'CONNECT'),
    pg_catalog.has_database_privilege(current_principal.rolname, pg_catalog.current_database(), 'CONNECT'),
    pg_catalog.has_schema_privilege(current_principal.rolname, 'platform', 'CREATE')
FROM pg_catalog.pg_roles AS session_principal
JOIN pg_catalog.pg_roles AS current_principal ON current_principal.rolname = current_user
LEFT JOIN pg_catalog.pg_auth_members AS membership
    ON membership.member = session_principal.oid AND membership.roleid = current_principal.oid
WHERE session_principal.rolname = session_user;
