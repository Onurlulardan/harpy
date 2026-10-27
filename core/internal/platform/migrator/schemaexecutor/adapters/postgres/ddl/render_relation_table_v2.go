package ddl

import (
	"fmt"
	"strconv"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func renderRelationTableV2(
	schemaName string,
	step schemaexecutor.ValidatedStep,
) (Statement, error) {
	mapping, err := v2Mapping(step.V2Mapping)
	if err != nil {
		return Statement{}, err
	}
	if step.Operation() == schemav1.DropRelationTable {
		arguments, ok := step.ArgumentValue().(*schemav2.ObjectArguments)
		if !ok {
			return Statement{}, fmt.Errorf("V2 relation drop arguments are invalid")
		}
		relation, err := identifierNameV2(mapping.Relations, arguments.ObjectID)
		if err != nil {
			return Statement{}, err
		}
		return Statement{Text: "DROP TABLE " + qualify(schemaName, relation)}, nil
	}
	arguments, ok := step.ArgumentValue().(*schemav2.RelationArguments)
	if !ok {
		return Statement{}, fmt.Errorf("V2 relation add arguments are invalid")
	}
	return renderRelationCreateV2(schemaName, mapping, *arguments)
}

func renderRelationCreateV2(
	schemaName string,
	mapping schemav2.IdentifierMapping,
	arguments schemav2.RelationArguments,
) (Statement, error) {
	relation, err := identifierNameV2(mapping.Relations, arguments.RelationID)
	if err != nil {
		return Statement{}, err
	}
	target, err := formNameV2(mapping, arguments.TargetFormID)
	if err != nil {
		return Statement{}, err
	}
	sourceConstraint, err := identifierNameV2(mapping.Constraints, arguments.SourceConstraintID)
	if err != nil {
		return Statement{}, err
	}
	targetConstraint, err := identifierNameV2(mapping.Constraints, arguments.TargetConstraintID)
	if err != nil {
		return Statement{}, err
	}
	deleteAction := "RESTRICT"
	if arguments.DeletePolicy == schemav2.DeleteCascade {
		deleteAction = "CASCADE"
	}
	text := "CREATE TABLE " + qualify(schemaName, relation) + " (" +
		"source_id uuid NOT NULL, target_id uuid NOT NULL, edge_slot smallint NOT NULL, " +
		"created_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(), " +
		"PRIMARY KEY (source_id, target_id), CONSTRAINT " + quote(sourceConstraint) +
		" FOREIGN KEY (source_id) REFERENCES " +
		qualify(schemaName, mapping.Form.PhysicalName) + " (id) ON DELETE " + deleteAction +
		", CONSTRAINT " + quote(targetConstraint) + " FOREIGN KEY (target_id) REFERENCES " +
		qualify(schemaName, target) + " (id) ON DELETE " + deleteAction
	if arguments.Cardinality == schemav2.CardinalityOne {
		text += ", CONSTRAINT " + quote(relation+"_edge_slot_check") +
			" CHECK (edge_slot = 1)"
		cardinality, err := identifierNameV2(mapping.Constraints, *arguments.CardinalityConstraintID)
		if err != nil {
			return Statement{}, err
		}
		text += ", CONSTRAINT " + quote(cardinality) + " UNIQUE (source_id)"
	} else {
		limit := strconv.Itoa(schemav2.MaximumRelationEdgesPerSource)
		text += ", CONSTRAINT " + quote(relation+"_edge_slot_check") +
			" CHECK (edge_slot BETWEEN 1 AND " + limit + ")" +
			", CONSTRAINT " + quote(relation+"_slot_key") +
			" UNIQUE (source_id, edge_slot)"
	}
	return Statement{Text: text + ")"}, nil
}
