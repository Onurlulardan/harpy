package ddl

import (
	"fmt"
	"strings"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func renderIndex(
	renderContext RenderContext,
	step schemaexecutor.ValidatedStep,
) (Statement, error) {
	if step.ContractVersion() == schemav2.PlanContractVersion {
		return renderIndexV2(renderContext, step)
	}
	schemaName := renderContext.TenantSchema
	switch step.Operation() {
	case schemav1.CreateIndex, schemav1.CreateIndexConcurrent:
		arguments, ok := step.ArgumentValue().(*schemav1.IndexArguments)
		if !ok {
			return Statement{}, fmt.Errorf("index create arguments are invalid")
		}
		return renderIndexCreate(schemaName, step.Mapping(), *arguments,
			step.Operation() == schemav1.CreateIndexConcurrent)
	case schemav1.CleanupInvalidIndex:
		arguments, ok := step.ArgumentValue().(*schemav1.IndexArguments)
		if !ok {
			return Statement{}, fmt.Errorf("index cleanup arguments are invalid")
		}
		name, err := identifierName(step.Mapping().Indexes, arguments.IndexID)
		if err != nil {
			return Statement{}, err
		}
		return Statement{Text: "DROP INDEX CONCURRENTLY " + qualify(schemaName, name)}, nil
	default:
		arguments, ok := step.ArgumentValue().(*schemav1.ObjectArguments)
		if !ok {
			return Statement{}, fmt.Errorf("index drop arguments are invalid")
		}
		name, err := identifierName(step.Mapping().Indexes, arguments.ObjectID)
		if err != nil {
			return Statement{}, err
		}
		concurrent := step.Operation() != schemav1.DropIndex
		keyword := ""
		if concurrent {
			keyword = " CONCURRENTLY"
		}
		return Statement{Text: "DROP INDEX" + keyword + " " + qualify(schemaName, name)}, nil
	}
}

func renderIndexCreate(
	schemaName string,
	mapping schemav1.IdentifierMapping,
	arguments schemav1.IndexArguments,
	concurrent bool,
) (Statement, error) {
	name, err := identifierName(mapping.Indexes, arguments.IndexID)
	if err != nil {
		return Statement{}, err
	}
	columns := make([]string, len(arguments.FieldIDs))
	for index, fieldID := range arguments.FieldIDs {
		field, err := fieldName(mapping, fieldID)
		if err != nil {
			return Statement{}, err
		}
		columns[index] = quote(field)
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
		qualify(schemaName, mapping.Form.PhysicalName) + " USING " +
		string(arguments.Method) + " (" + strings.Join(columns, ", ") + ")"
	if arguments.ActiveRecordsOnly {
		text += " WHERE deleted_at IS NULL"
	}
	return Statement{Text: text}, nil
}
