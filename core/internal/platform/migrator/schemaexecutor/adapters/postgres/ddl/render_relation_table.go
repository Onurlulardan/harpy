package ddl

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func renderRelationTable(
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (Statement, error) {
	if step.ContractVersion() == schemav2.PlanContractVersion {
		return renderRelationTableV2(schemaName, step)
	}
	mapping := step.Mapping()
	if step.Operation() == schemav1.DropRelationTable {
		arguments, ok := step.ArgumentValue().(*schemav1.ObjectArguments)
		if !ok {
			return Statement{}, fmt.Errorf("relation drop arguments are invalid")
		}
		relation, err := identifierName(mapping.Relations, arguments.ObjectID)
		if err != nil {
			return Statement{}, err
		}
		return Statement{Text: "DROP TABLE " + qualify(schemaName, relation)}, nil
	}
	arguments, ok := step.ArgumentValue().(*schemav1.RelationArguments)
	if !ok {
		return Statement{}, fmt.Errorf("relation add arguments are invalid")
	}
	return renderRelationCreate(schemaName, mapping, *arguments)
}

func renderRelationCreate(
	schemaName string,
	mapping schemav1.IdentifierMapping,
	arguments schemav1.RelationArguments,
) (Statement, error) {
	relation, err := identifierName(mapping.Relations, arguments.RelationID)
	if err != nil {
		return Statement{}, err
	}
	target, err := formName(mapping, arguments.TargetFormID)
	if err != nil {
		return Statement{}, err
	}
	sourceConstraint, err := identifierName(mapping.Constraints, arguments.SourceConstraintID)
	if err != nil {
		return Statement{}, err
	}
	targetConstraint, err := identifierName(mapping.Constraints, arguments.TargetConstraintID)
	if err != nil {
		return Statement{}, err
	}
	deleteAction := "RESTRICT"
	if arguments.DeletePolicy == schemav1.DeleteCascade {
		deleteAction = "CASCADE"
	}
	text := "CREATE TABLE " + qualify(schemaName, relation) + " (" +
		"source_id uuid NOT NULL, target_id uuid NOT NULL, " +
		"created_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(), " +
		"PRIMARY KEY (source_id, target_id), CONSTRAINT " + quote(sourceConstraint) +
		" FOREIGN KEY (source_id) REFERENCES " +
		qualify(schemaName, mapping.Form.PhysicalName) + " (id) ON DELETE " + deleteAction +
		", CONSTRAINT " + quote(targetConstraint) + " FOREIGN KEY (target_id) REFERENCES " +
		qualify(schemaName, target) + " (id) ON DELETE " + deleteAction
	if arguments.CardinalityConstraintID != nil {
		cardinality, err := identifierName(mapping.Constraints, *arguments.CardinalityConstraintID)
		if err != nil {
			return Statement{}, err
		}
		text += ", CONSTRAINT " + quote(cardinality) + " UNIQUE (source_id)"
	}
	return Statement{Text: text + ")"}, nil
}
