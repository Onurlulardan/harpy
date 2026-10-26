package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func (statement IncludeTargetStatement) ValidatePlan(
	plan resolved.AuthorizedPlan, graph IncludeGraph, node IncludeNode,
) error {
	if statement.Validate() != nil ||
		statement.binding.validatePlan(plan, graph, node) != nil ||
		statement.catalog != node.targetCatalog ||
		statement.stableFieldID != mustFieldID(node.targetStable) {
		return ErrContract
	}
	return nil
}

func mustFieldID(field api.FieldDefinition) uuid.UUID {
	identifier, err := field.ID()
	if err != nil {
		return uuid.Nil
	}
	return identifier
}
