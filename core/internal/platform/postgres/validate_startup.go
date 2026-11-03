package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func validateStartup(
	ctx context.Context,
	connection *pgx.Conn,
	profile Profile,
	contract profileContract,
) (StartupResult, error) {
	catalog, err := readCatalogPreflight(ctx, connection)
	if err != nil {
		return StartupResult{}, err
	}
	origin, err := originForVectorSchema(catalog.vectorSchema)
	if err != nil {
		return StartupResult{}, err
	}
	if err := validateRolePreflight(ctx, connection, contract); err != nil {
		return StartupResult{}, err
	}
	if err := validateSession(ctx, connection, contract); err != nil {
		return StartupResult{}, fmt.Errorf("%w: session contract", ErrStartupPreflight)
	}
	return StartupResult{
		profile: profile, applicationName: contract.applicationName,
		serverVersion: catalog.serverVersion, databaseOID: uint32(catalog.databaseOID),
		databaseName: catalog.databaseName, systemIdentifier: catalog.systemIdentifier,
		vectorSchema: catalog.vectorSchema, databaseOrigin: origin,
	}, nil
}
