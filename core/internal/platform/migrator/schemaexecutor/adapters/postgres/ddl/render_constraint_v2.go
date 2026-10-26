package ddl

import (
	"fmt"
	"strings"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func renderConstraintV2(
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (Statement, error) {
	mapping, err := v2Mapping(step.V2Mapping)
	if err != nil {
		return Statement{}, err
	}
	table := qualify(schemaName, mapping.Form.PhysicalName)
	if step.Operation() == schemav1.DropConstraint {
		arguments, ok := step.ArgumentValue().(*schemav2.ObjectArguments)
		if !ok {
			return Statement{}, fmt.Errorf("V2 constraint object arguments are invalid")
		}
		if arguments.Generation == nil {
			return Statement{}, fmt.Errorf("V2 constraint generation is missing")
		}
		name, err := constraintNameV2(
			mapping.Constraints, arguments.ObjectID, *arguments.Generation,
		)
		if err != nil {
			return Statement{}, err
		}
		return Statement{Text: "ALTER TABLE " + table +
			" DROP CONSTRAINT " + quote(name)}, nil
	}
	arguments, ok := step.ArgumentValue().(*schemav2.ConstraintArguments)
	if !ok {
		return Statement{}, fmt.Errorf("V2 constraint add arguments are invalid")
	}
	name, err := constraintNameV2(
		mapping.Constraints, arguments.ConstraintID, arguments.Generation,
	)
	if err != nil {
		return Statement{}, err
	}
	if step.Operation() == schemav1.ValidateConstraint {
		return Statement{Text: "ALTER TABLE " + table +
			" VALIDATE CONSTRAINT " + quote(name)}, nil
	}
	columns := make([]string, len(arguments.Fields))
	for index, field := range arguments.Fields {
		name, fieldErr := fieldNameV2(mapping, field)
		if fieldErr != nil {
			return Statement{}, fieldErr
		}
		columns[index] = quote(name)
	}
	prefix := "ALTER TABLE " + table + " ADD CONSTRAINT " + quote(name) + " "
	switch arguments.Kind {
	case schemav2.ConstraintUnique:
		return Statement{Text: prefix + "UNIQUE (" + strings.Join(columns, ", ") + ")"}, nil
	case schemav2.ConstraintNotNull:
		return Statement{Text: prefix + "CHECK (" + columns[0] + " IS NOT NULL) NOT VALID"}, nil
	case schemav2.ConstraintTextLength:
		return Statement{Text: fmt.Sprintf(
			"%sCHECK (pg_catalog.char_length(%s) BETWEEN %d AND %d) NOT VALID",
			prefix, columns[0], *arguments.Minimum, *arguments.Maximum,
		)}, nil
	default:
		return Statement{}, fmt.Errorf("V2 constraint kind is not renderable")
	}
}
