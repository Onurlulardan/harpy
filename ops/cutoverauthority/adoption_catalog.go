package main

import (
	"context"
	"fmt"
	"strings"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
)

type catalogState struct {
	oid        uint32
	owner      string
	definition string
	acl        string
}

func lockAdoptionCatalog(
	ctx context.Context, transaction pgx.Tx,
	operation cutoverv1.AuthorityAdoptionOperation,
) error {
	if _, err := transaction.Exec(ctx, `SELECT pg_catalog.pg_advisory_xact_lock(
pg_catalog.hashtextextended('harpy.cutover_authority.adoption',0))`); err != nil {
		return fmt.Errorf("authority adoption lock failed")
	}
	protected := make([]string, 0, len(operation.Objects))
	for _, object := range operation.Objects {
		if object.Kind == "table" {
			protected = append(protected,
				pgx.Identifier{object.SchemaName, object.ObjectName}.Sanitize())
		}
	}
	if err := lockTables(ctx, transaction, protected, "ACCESS EXCLUSIVE"); err != nil {
		return err
	}
	dependencies := make([]string, 0, len(operation.Dependencies))
	for _, dependency := range operation.Dependencies {
		if dependency.Kind == "table" {
			dependencies = append(dependencies,
				pgx.Identifier{dependency.SchemaName, dependency.ObjectName}.Sanitize())
		}
	}
	return lockTables(ctx, transaction, dependencies, "SHARE")
}

func lockTables(
	ctx context.Context, transaction pgx.Tx, tables []string, mode string,
) error {
	if len(tables) == 0 {
		return nil
	}
	statement := "LOCK TABLE ONLY " + strings.Join(tables, ",") + " IN " + mode + " MODE"
	if _, err := transaction.Exec(ctx, statement); err != nil {
		return fmt.Errorf("authority catalog lock failed")
	}
	return nil
}

func readObjectStates(
	ctx context.Context, query catalogQuerier,
	objects []cutoverv1.AuthorityAdoptionObject,
) ([]catalogState, error) {
	states := make([]catalogState, len(objects))
	for index, object := range objects {
		state, err := readCatalogState(ctx, query, object.Kind, object.SchemaName,
			object.ObjectName, object.IdentityArguments)
		if err != nil {
			return nil, err
		}
		states[index] = state
	}
	return states, nil
}

func readDependencyStates(
	ctx context.Context, query catalogQuerier,
	dependencies []cutoverv1.AuthorityCatalogDependency,
) ([]catalogState, error) {
	states := make([]catalogState, len(dependencies))
	for index, dependency := range dependencies {
		state, err := readCatalogState(ctx, query, dependency.Kind,
			dependency.SchemaName, dependency.ObjectName, dependency.IdentityArguments)
		if err != nil {
			return nil, err
		}
		states[index] = state
	}
	return states, nil
}

type catalogQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func readCatalogState(
	ctx context.Context, query catalogQuerier,
	kind, schema, name, arguments string,
) (catalogState, error) {
	statement, ok := catalogLocator(kind)
	if !ok {
		return catalogState{}, fmt.Errorf("authority catalog kind is unsupported")
	}
	var state catalogState
	if err := query.QueryRow(ctx, statement, schema, name, arguments).Scan(
		&state.oid, &state.owner,
	); err != nil || state.oid == 0 {
		return catalogState{}, fmt.Errorf("authority catalog object is missing")
	}
	if err := query.QueryRow(ctx, `SELECT
cutover_authority.catalog_definition_sha256_v1($1::text,$2::oid),
cutover_authority.catalog_acl_sha256_v1($1::text,$2::oid)`, kind, state.oid).Scan(
		&state.definition, &state.acl,
	); err != nil {
		return catalogState{}, fmt.Errorf("authority catalog fingerprint failed")
	}
	return state, nil
}

func catalogLocator(kind string) (string, bool) {
	switch kind {
	case "schema":
		return `SELECT namespace.oid, pg_catalog.pg_get_userbyid(namespace.nspowner)
FROM pg_catalog.pg_namespace AS namespace
WHERE namespace.nspname=$1 AND $2=$1 AND $3=''`, true
	case "table":
		return `SELECT relation.oid, pg_catalog.pg_get_userbyid(relation.relowner)
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
WHERE namespace.nspname=$1 AND relation.relname=$2
  AND relation.relkind IN ('r','p') AND $3=''`, true
	case "function":
		return `SELECT procedure_row.oid, pg_catalog.pg_get_userbyid(procedure_row.proowner)
FROM pg_catalog.pg_proc AS procedure_row
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=procedure_row.pronamespace
WHERE namespace.nspname=$1 AND procedure_row.proname=$2
  AND pg_catalog.replace(pg_catalog.oidvectortypes(
      procedure_row.proargtypes),' ','')=$3`, true
	default:
		return "", false
	}
}

func transferAdoptedObjects(
	ctx context.Context, transaction pgx.Tx,
	objects []cutoverv1.AuthorityAdoptionObject,
) error {
	for _, object := range objects {
		if object.Action != "adopt" {
			continue
		}
		qualified := pgx.Identifier{object.SchemaName, object.ObjectName}.Sanitize()
		owner := pgx.Identifier{object.AdoptedOwner}.Sanitize()
		var statement string
		switch object.Kind {
		case "schema":
			statement = "ALTER SCHEMA " + pgx.Identifier{object.SchemaName}.Sanitize() +
				" OWNER TO " + owner
		case "table":
			statement = "ALTER TABLE " + qualified + " OWNER TO " + owner
		case "function":
			identity, err := resolveFunctionTransferIdentity(ctx, transaction, object)
			if err != nil {
				return err
			}
			statement = "ALTER FUNCTION " + identity + " OWNER TO " + owner
		default:
			return fmt.Errorf("authority adoption object kind is unsupported")
		}
		if _, err := transaction.Exec(ctx, statement); err != nil {
			return fmt.Errorf("authority object ownership transfer failed")
		}
	}
	return nil
}
