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
