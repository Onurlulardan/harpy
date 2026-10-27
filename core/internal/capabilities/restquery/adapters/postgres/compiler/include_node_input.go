package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type includeNodeInput struct {
	include        resolved.Include
	sourceCatalog  api.CatalogReference
	sourceStable   api.FieldDefinition
	depth          int
	ordinal        int
	parentOrdinal  int
	siblingOrdinal int
}

func newIncludeNode(
	input includeNodeInput, binding includePlanBinding,
) (IncludeNode, error) {
	if input.include.Validate() != nil || input.sourceCatalog.Validate() != nil ||
		!validIncludeStableField(input.sourceStable) || binding.validate() != nil {
		return IncludeNode{}, ErrContract
	}
	relation, relationErr := input.include.Relation()
	target, targetErr := input.include.Target()
	targetStorage, targetStorageErr := input.include.TargetStorage()
	relationStorage, relationStorageErr := input.include.RelationStorage()
	targetStable, stableErr := input.include.StableID()
	path, pathErr := input.include.Path()
	limit, limitErr := input.include.Limit()
	relationName, nameErr := relation.PublicName()
	cardinality, cardinalityErr := relation.Cardinality()
	if relationErr != nil || targetErr != nil || targetStorageErr != nil ||
		relationStorageErr != nil || stableErr != nil || pathErr != nil ||
		limitErr != nil || nameErr != nil || cardinalityErr != nil {
		return IncludeNode{}, ErrContract
	}
	node := IncludeNode{
		sourceCatalog: input.sourceCatalog, targetCatalog: target,
		targetStorage: targetStorage, relationStorage: relationStorage,
		sourceStable: input.sourceStable, targetStable: targetStable,
		path: path, relationName: relationName, cardinality: cardinality,
		limit: limit, depth: input.depth, ordinal: input.ordinal,
		parentOrdinal: input.parentOrdinal, siblingOrdinal: input.siblingOrdinal,
		binding: binding, validated: true,
	}
	seal, err := includeNodeSeal(node)
	if err != nil {
		return IncludeNode{}, ErrContract
	}
	node.stateSeal = seal
	if node.Validate() != nil {
		return IncludeNode{}, ErrContract
	}
	return node, nil
}
