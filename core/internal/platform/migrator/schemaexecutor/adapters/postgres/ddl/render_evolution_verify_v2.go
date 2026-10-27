package ddl

import (
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/jackc/pgx/v5"
)

func RenderEvolutionVerificationV2(
	schemaName, tableName, sourceName, targetName string,
	converter schemav2.Converter,
	defaultLiteral *schemav2.TypedLiteral,
) (string, error) {
	table := pgx.Identifier{schemaName, tableName}.Sanitize()
	source := "record." + pgx.Identifier{sourceName}.Sanitize()
	target := "record." + pgx.Identifier{targetName}.Sanitize()
	forward, reverse, err := evolutionConversionPair(converter, source, target)
	if err != nil {
		return "", err
	}
	effectiveForward := forward
	if defaultLiteral != nil {
		literal, literalErr := RenderTypedLiteralV2(*defaultLiteral)
		if literalErr != nil {
			return "", literalErr
		}
		effectiveForward = "COALESCE(" + forward + "," + literal + ")"
	}
	sourceCanonical := "COALESCE(pg_catalog.to_jsonb(" + effectiveForward + ")::text,'null')"
	targetCanonical := "COALESCE(pg_catalog.to_jsonb(" + target + ")::text,'null')"
	sourceHash := "record.id::text||E'\\x1f'||" + sourceCanonical
	targetHash := "record.id::text||E'\\x1f'||" + targetCanonical
	return "SELECT " +
		"pg_catalog.count(*) FILTER (WHERE (" + effectiveForward + " IS NULL)<>(" + target + " IS NULL))," +
		"pg_catalog.count(*) FILTER (WHERE NOT (" + target + " IS NOT DISTINCT FROM " + effectiveForward + "))," +
		"pg_catalog.count(*) FILTER (WHERE CASE WHEN " + source + " IS NULL THEN FALSE " +
		"ELSE NOT (" + source + " IS NOT DISTINCT FROM " + reverse + ") END)," +
		"COALESCE(pg_catalog.bit_xor(pg_catalog.hashtextextended(" + sourceHash + ",6101)),0)," +
		"COALESCE(pg_catalog.bit_xor(pg_catalog.hashtextextended(" + targetHash + ",6101)),0)," +
		"COALESCE(pg_catalog.bit_xor(pg_catalog.hashtextextended(" + sourceHash + ",6102)),0)," +
		"COALESCE(pg_catalog.bit_xor(pg_catalog.hashtextextended(" + targetHash + ",6102)),0) " +
		"FROM " + table + " AS record", nil
}

func evolutionConversionPair(
	converter schemav2.Converter,
	source, target string,
) (string, string, error) {
	switch converter {
	case schemav2.ConvertIdentity, schemav2.ConvertDecimalWiden:
		return source, target, nil
	case schemav2.ConvertIntegerDecimal:
		return source + "::numeric", target + "::bigint", nil
	case schemav2.ConvertIntegerText:
		return source + "::text", target + "::bigint", nil
	case schemav2.ConvertDecimalText:
		return source + "::text", target + "::numeric", nil
	case schemav2.ConvertDateTimestamp:
		return source + "::timestamp AT TIME ZONE 'UTC'",
			"(" + target + " AT TIME ZONE 'UTC')::date", nil
	case schemav2.ConvertUUIDText:
		return source + "::text", target + "::uuid", nil
	default:
		return "", "", fmt.Errorf("evolution converter is not verifiable")
	}
}
