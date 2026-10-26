package ddl

import (
	"fmt"

	pagev1 "github.com/Onurlulardan/harpy/contracts/formsrecordpage/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func renderBackfillV2(
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (Statement, error) {
	mapping, err := v2Mapping(step.V2Mapping)
	if err != nil {
		return Statement{}, err
	}
	arguments, ok := step.ArgumentValue().(*schemav2.BackfillArguments)
	if !ok {
		return Statement{}, fmt.Errorf("V2 backfill arguments are invalid")
	}
	source, err := fieldNameV2(mapping, arguments.Source)
	if err != nil {
		return Statement{}, err
	}
	target, err := fieldNameV2(mapping, arguments.Target)
	if err != nil {
		return Statement{}, err
	}
	conversion, err := renderConversionV2(arguments.Converter, "selected.source_value")
	if err != nil {
		return Statement{}, err
	}
	if arguments.DefaultLiteral != nil {
		literal, literalErr := RenderTypedLiteralV2(*arguments.DefaultLiteral)
		if literalErr != nil {
			return Statement{}, literalErr
		}
		conversion = "COALESCE(" + conversion + ", " + literal + ")"
	}
	text := renderBackfillV2PageSQL(qualify(schemaName, mapping.Form.PhysicalName),
		quote(source), quote(target), conversion)
	return Statement{Text: text, RecordPageEvidence: true}, nil
}

func renderBackfillV2PageSQL(
	table, source, target, conversion string,
) string {
	rowDocument := "pg_catalog.convert_to(pg_catalog.jsonb_build_array(" +
		"id::text,row_version,pg_catalog.to_jsonb(" + source + "))::text,'UTF8')"
	separator := "pg_catalog.convert_to(E'\\n','UTF8')"
	page := func(name string) string {
		return "COALESCE(pg_catalog.string_agg(" + name + ".row_payload," + separator +
			" ORDER BY " + name + ".id),''::bytea)"
	}
	return fmt.Sprintf("WITH candidates AS MATERIALIZED ("+
		"SELECT id,row_version,%s AS source_value,%s AS row_payload FROM %s "+
		"WHERE ($1::uuid IS NULL OR id>$1::uuid) ORDER BY id LIMIT %d FOR UPDATE),"+
		"ranked AS MATERIALIZED (SELECT candidates.*,"+
		"pg_catalog.row_number() OVER (ORDER BY id) AS position,"+
		"pg_catalog.octet_length(row_payload) AS row_bytes FROM candidates),"+
		"measured AS MATERIALIZED (SELECT ranked.*,"+
		"pg_catalog.sum(row_bytes+CASE WHEN position>1 THEN 1 ELSE 0 END) "+
		"OVER (ORDER BY id) AS cumulative_bytes FROM ranked),"+
		"boundary AS (SELECT pg_catalog.min(position) AS position FROM measured "+
		"WHERE row_bytes>%d OR position>%d OR cumulative_bytes>%d),"+
		"selected AS MATERIALIZED (SELECT measured.* FROM measured CROSS JOIN boundary "+
		"WHERE measured.position<COALESCE(boundary.position,%d) ORDER BY measured.id),"+
		"updated AS (UPDATE %s AS target SET %s=%s,"+
		"updated_at=pg_catalog.statement_timestamp(),row_version=target.row_version+1 "+
		"FROM selected WHERE target.id=selected.id AND target.row_version=selected.row_version "+
		"RETURNING target.id,selected.row_payload) "+
		"SELECT (SELECT id::text FROM selected ORDER BY id DESC LIMIT 1),"+
		"(SELECT pg_catalog.count(*)::bigint FROM selected),"+
		"pg_catalog.octet_length((SELECT %s FROM selected))::bigint,"+
		"pg_catalog.encode(pg_catalog.sha256((SELECT %s FROM selected)),'hex'),"+
		"(SELECT id::text FROM updated ORDER BY id DESC LIMIT 1),"+
		"(SELECT pg_catalog.count(*)::bigint FROM updated),"+
		"pg_catalog.octet_length((SELECT %s FROM updated))::bigint,"+
		"pg_catalog.encode(pg_catalog.sha256((SELECT %s FROM updated)),'hex'),"+
		"COALESCE((SELECT row_bytes FROM measured,boundary "+
		"WHERE measured.position=boundary.position),0)::bigint",
		source, rowDocument, table, pagev1.MaximumRows+1,
		pagev1.MaximumRowBytes, pagev1.MaximumRows, pagev1.MaximumCanonicalBytes,
		pagev1.MaximumRows+1, table, target, conversion,
		page("selected"), page("selected"), page("updated"), page("updated"))
}

func renderConversionV2(converter schemav2.Converter, value string) (string, error) {
	switch converter {
	case schemav2.ConvertIdentity, schemav2.ConvertDecimalWiden:
		return value, nil
	case schemav2.ConvertIntegerDecimal:
		return value + "::numeric", nil
	case schemav2.ConvertIntegerText, schemav2.ConvertDecimalText, schemav2.ConvertUUIDText:
		return value + "::text", nil
	case schemav2.ConvertDateTimestamp:
		return value + "::timestamp AT TIME ZONE 'UTC'", nil
	default:
		return "", fmt.Errorf("V2 converter is not renderable")
	}
}
