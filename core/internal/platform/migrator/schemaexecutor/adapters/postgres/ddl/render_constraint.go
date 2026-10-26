package ddl

import (
	"fmt"
	"strings"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func renderConstraint(
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (Statement, error) {
	if step.ContractVersion() == schemav2.PlanContractVersion {
		return renderConstraintV2(schemaName, step)
	}
	mapping := step.Mapping()
	table := qualify(schemaName, mapping.Form.PhysicalName)
	if step.Operation() == schemav1.DropConstraint {
		arguments, ok := step.ArgumentValue().(*schemav1.ObjectArguments)
		if !ok {
			return Statement{}, fmt.Errorf("constraint object arguments are invalid")
		}
		name, err := identifierName(mapping.Constraints, arguments.ObjectID)
		if err != nil {
			return Statement{}, err
		}
		return Statement{Text: "ALTER TABLE " + table +
			" DROP CONSTRAINT " + quote(name)}, nil
	}
	arguments, ok := step.ArgumentValue().(*schemav1.ConstraintArguments)
	if !ok {
		return Statement{}, fmt.Errorf("constraint add arguments are invalid")
	}
	name, err := identifierName(mapping.Constraints, arguments.ConstraintID)
	if err != nil {
		return Statement{}, err
	}
	if step.Operation() == schemav1.ValidateConstraint {
		return Statement{Text: "ALTER TABLE " + table +
			" VALIDATE CONSTRAINT " + quote(name)}, nil
	}
	columns := make([]string, len(arguments.FieldIDs))
	for index, fieldID := range arguments.FieldIDs {
		field, err := fieldName(mapping, fieldID)
		if err != nil {
			return Statement{}, err
		}
		columns[index] = quote(field)
	}
	prefix := "ALTER TABLE " + table + " ADD CONSTRAINT " + quote(name) + " "
	switch arguments.Kind {
	case schemav1.ConstraintUnique:
		return Statement{Text: prefix + "UNIQUE (" + strings.Join(columns, ", ") + ")"}, nil
	case schemav1.ConstraintNotNull:
		return Statement{Text: prefix + "CHECK (" + columns[0] + " IS NOT NULL) NOT VALID"}, nil
	case schemav1.ConstraintTextLength:
		return Statement{Text: fmt.Sprintf(
			"%sCHECK (pg_catalog.char_length(%s) BETWEEN %d AND %d) NOT VALID",
			prefix, columns[0], *arguments.Minimum, *arguments.Maximum,
		)}, nil
	default:
		return Statement{}, fmt.Errorf("constraint kind is not renderable")
	}
}
