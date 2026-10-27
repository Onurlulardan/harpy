package main

import (
	"context"
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
)

func assertAuthorityNamespace(
	ctx context.Context, query catalogQuerier, expected int,
) error {
	var count int
	err := query.QueryRow(ctx, `SELECT 1 +
    (SELECT count(*) FROM pg_catalog.pg_class AS relation
     JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
     WHERE namespace.nspname='cutover_authority' AND relation.relkind IN ('r','p')) +
    (SELECT count(*) FROM pg_catalog.pg_proc AS procedure_row
     JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=procedure_row.pronamespace
     WHERE namespace.nspname='cutover_authority')`).Scan(&count)
	if err != nil || count != expected {
		return fmt.Errorf("protected authority namespace differs")
	}
	return nil
}

func assertPreAdoptionStates(
	states []catalogState, objects []cutoverv1.AuthorityAdoptionObject,
) error {
	if len(states) != len(objects) {
		return fmt.Errorf("pre-adoption object set differs")
	}
	for index, state := range states {
		object := objects[index]
		if state.owner != object.PreAdoptionOwner ||
			state.definition != object.PreDefinitionSHA256 ||
			state.acl != object.PreAdoptionACLSHA256 {
			return fmt.Errorf("pre-adoption catalog differs")
		}
	}
	return nil
}

func assertPostAdoptionStates(
	states []catalogState, objects []cutoverv1.AuthorityAdoptionObject,
) error {
	if len(states) != len(objects) {
		return fmt.Errorf("post-adoption object set differs")
	}
	for index, state := range states {
		object := objects[index]
		if state.owner != object.AdoptedOwner ||
			state.definition != object.AdoptedDefinitionSHA256 ||
			state.acl != object.AdoptedACLSHA256 {
			return fmt.Errorf("post-adoption catalog differs")
		}
	}
	return nil
}

func assertDependencyStates(
	states []catalogState, dependencies []cutoverv1.AuthorityCatalogDependency,
) error {
	if len(states) != len(dependencies) {
		return fmt.Errorf("authority dependency set differs")
	}
	for index, state := range states {
		dependency := dependencies[index]
		if state.owner != dependency.Owner || state.definition != dependency.DefinitionSHA256 ||
			state.acl != dependency.ACLSHA256 {
			return fmt.Errorf("authority dependency catalog differs")
		}
	}
	return nil
}

func assertDependencyStatesUnchanged(before, after []catalogState) error {
	if len(before) != len(after) {
		return fmt.Errorf("authority dependency set changed")
	}
	for index := range before {
		if before[index] != after[index] {
			return fmt.Errorf("authority dependency changed during adoption")
		}
	}
	return nil
}

func assertAuthorityOwner(
	ctx context.Context, query catalogQuerier, expectedObjects int,
) error {
	var exact bool
	err := query.QueryRow(ctx, `SELECT count(*)=1 AND bool_and(
    NOT role.rolcanlogin AND NOT role.rolinherit AND NOT role.rolsuper
    AND NOT role.rolcreatedb AND NOT role.rolcreaterole AND NOT role.rolreplication
    AND NOT role.rolbypassrls AND role.rolconnlimit=-1 AND role.rolpassword IS NULL
    AND (role.rolvaliduntil IS NULL OR role.rolvaliduntil='infinity'::timestamptz)
    AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
                    WHERE membership.roleid=role.oid OR membership.member=role.oid))
FROM pg_catalog.pg_authid AS role
WHERE role.rolname='harpy_cutover_authority_owner'`).Scan(&exact)
	if err != nil || !exact {
		return fmt.Errorf("cutover authority owner role differs")
	}
	return assertAuthorityNamespace(ctx, query, expectedObjects)
}

func restoreMigratorAfterFailure(
	connection *pgx.Conn, policy cutoverv1.AuthorityQuiescedRolePolicy,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	observed, err := readQuiescedRole(ctx, connection, policy.Principal)
	if err != nil {
		return fmt.Errorf("migrator recovery observation failed")
	}
	if observed.state == policy.PostState {
		return nil
	}
	if observed.state != policy.TemporaryState {
		return fmt.Errorf("migrator recovery state differs")
	}
	if err := restoreMigrator(ctx, connection, policy); err != nil {
		return err
	}
	restored, err := readQuiescedRole(ctx, connection, policy.Principal)
	if err != nil || restored.state != policy.PostState ||
		restored.verifier != observed.verifier {
		return fmt.Errorf("migrator recovery verification failed")
	}
	return nil
}
