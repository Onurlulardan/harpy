package postgres

import (
	"context"
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

type catalogPreflight struct {
	serverVersion    int
	databaseOID      int64
	databaseName     string
	databaseOwner    string
	encoding         string
	localeProvider   string
	collation        string
	characterType    string
	checksums        string
	inRecovery       bool
	vectorVersion    string
	vectorSchema     string
	vectorOwner      string
	systemIdentifier string
}

func readCatalogPreflight(
	ctx context.Context,
	queryer roleQueryer,
) (catalogPreflight, error) {
	var result catalogPreflight
	err := queryer.QueryRow(ctx, preflightCatalogSQL).Scan(
		&result.serverVersion, &result.databaseOID, &result.databaseName,
		&result.databaseOwner, &result.encoding, &result.localeProvider,
		&result.collation, &result.characterType, &result.checksums,
		&result.inRecovery, &result.vectorVersion, &result.vectorSchema,
		&result.vectorOwner, &result.systemIdentifier,
	)
	if err != nil {
		return catalogPreflight{}, fmt.Errorf("%w: catalog query", ErrStartupPreflight)
	}
	if result.serverVersion != runtimev1.PostgresServerVersionNumber ||
		result.databaseOID <= 0 || result.databaseOID > int64(^uint32(0)) ||
		result.databaseName == "" || result.databaseOwner != runtimev1.DatabaseOwnerRole ||
		result.encoding != runtimev1.PostgresEncoding ||
		result.localeProvider != runtimev1.PostgresLocaleProvider ||
		result.collation != runtimev1.PostgresCollation ||
		result.characterType != runtimev1.PostgresCharacterType ||
		result.checksums != "on" || result.inRecovery ||
		result.vectorVersion != runtimev1.VectorExtensionVersion ||
		result.vectorOwner == "" || result.systemIdentifier == "" {
		return catalogPreflight{}, fmt.Errorf("%w: catalog contract", ErrStartupPreflight)
	}
	if _, err := originForVectorSchema(result.vectorSchema); err != nil {
		return catalogPreflight{}, err
	}
	return result, nil
}
