package ddl

import (
	"fmt"
	"strings"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func renderIndexV2(
	renderContext RenderContext,
	step schemaexecutor.ValidatedStep,
) (Statement, error) {
	mapping, err := v2Mapping(step.V2Mapping)
	if err != nil {
		return Statement{}, err
	}
	switch step.Operation() {
	case schemav1.CreateIndex, schemav1.CreateIndexConcurrent:
		arguments, ok := step.ArgumentValue().(*schemav2.IndexArguments)
		if !ok {
			return Statement{}, fmt.Errorf("V2 index create arguments are invalid")
		}
		return renderIndexCreateV2(renderContext, mapping, *arguments,
			step.Operation() == schemav1.CreateIndexConcurrent)
	case schemav1.CleanupInvalidIndex:
		arguments, ok := step.ArgumentValue().(*schemav2.IndexArguments)
		if !ok {
			return Statement{}, fmt.Errorf("V2 index cleanup arguments are invalid")
		}
		name, err := indexNameV2(
			mapping.Indexes, arguments.IndexID, arguments.Generation,
		)
		if err != nil {
			return Statement{}, err
		}
		return Statement{Text: "DROP INDEX CONCURRENTLY " +
			qualify(renderContext.TenantSchema, name)}, nil
	default:
		arguments, ok := step.ArgumentValue().(*schemav2.ObjectArguments)
		if !ok {
			return Statement{}, fmt.Errorf("V2 index drop arguments are invalid")
		}
		if arguments.Generation == nil {
			return Statement{}, fmt.Errorf("V2 index drop generation is missing")
		}
		name, err := indexNameV2(
			mapping.Indexes, arguments.ObjectID, *arguments.Generation,
		)
		if err != nil {
			return Statement{}, err
		}
		keyword := ""
		if step.Operation() != schemav1.DropIndex {
			keyword = " CONCURRENTLY"
		}
		return Statement{Text: "DROP INDEX" + keyword + " " +
			qualify(renderContext.TenantSchema, name)}, nil
	}
}

func renderIndexCreateV2(
	renderContext RenderContext,
	mapping schemav2.IdentifierMapping,
	arguments schemav2.IndexArguments,
	concurrent bool,
) (Statement, error) {
	name, err := indexNameV2(mapping.Indexes, arguments.IndexID, arguments.Generation)
	if err != nil {
		return Statement{}, err
	}
	columns := make([]string, len(arguments.Fields))
	for index, field := range arguments.Fields {
		name, fieldErr := fieldNameV2(mapping, schemav2.FieldReference{
			FieldID: field.FieldID, Generation: field.Generation, Variant: field.Variant,
		})
		if fieldErr != nil {
			return Statement{}, fieldErr
		}
		columns[index] = quote(name)
		if arguments.OperatorClass == schemav2.OperatorClassGINTrigram {
			if renderContext.TrigramSchema == "" {
				return Statement{}, fmt.Errorf("trigram extension schema is unavailable")
			}
			columns[index] += " " + qualify(
				renderContext.TrigramSchema, string(schemav2.OperatorClassGINTrigram),
			)
		}
	}
	prefix := "CREATE "
	if arguments.Unique {
		prefix += "UNIQUE "
	}
	prefix += "INDEX "
	if concurrent {
		prefix += "CONCURRENTLY "
	}
	text := prefix + quote(name) + " ON " +
		qualify(renderContext.TenantSchema, mapping.Form.PhysicalName) + " USING " +
		string(arguments.Method) + " (" + strings.Join(columns, ", ") + ")"
	if arguments.ActiveRecordsOnly {
		text += " WHERE deleted_at IS NULL"
	}
	return Statement{Text: text}, nil
}
