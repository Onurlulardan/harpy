package postgres

import (
	"context"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
)

func schemaExecutorVerifyEvolutionDataV2(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	mapping schemav2.IdentifierMapping,
	contracts []schemav2.EvolutionDataContract,
) error {
	for _, contract := range contracts {
		source, err := schemaExecutorFieldNameV2(mapping, contract.Source)
		if err != nil {
			return err
		}
		target, err := schemaExecutorFieldNameV2(mapping, contract.Target)
		if err != nil {
			return err
		}
		statement, err := ddl.RenderEvolutionVerificationV2(
			schemaName, mapping.Form.PhysicalName, source, target,
			contract.Converter, contract.DefaultLiteral,
		)
		if err != nil {
			return err
		}
		var nullMismatch, forwardMismatch, roundTripMismatch int64
		var sourceFirst, targetFirst, sourceSecond, targetSecond int64
		err = querier.QueryRow(ctx, statement).Scan(
			&nullMismatch, &forwardMismatch, &roundTripMismatch,
			&sourceFirst, &targetFirst, &sourceSecond, &targetSecond,
		)
		if err != nil {
			return schemaExecutorBlocked("schema_evolution_data_unreadable", nil)
		}
		if nullMismatch != 0 || forwardMismatch != 0 || roundTripMismatch != 0 ||
			sourceFirst != targetFirst || sourceSecond != targetSecond {
			return schemaExecutorBlocked("schema_evolution_data_mismatch", nil)
		}
	}
	return nil
}
