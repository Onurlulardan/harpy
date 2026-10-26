package compiler

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type IncludeNode struct {
	sourceCatalog   api.CatalogReference
	targetCatalog   api.CatalogReference
	targetStorage   api.ResourceStorage
	relationStorage api.RelationStorage
	sourceStable    api.FieldDefinition
	targetStable    api.FieldDefinition
	path            resolved.RequestPath
	relationName    string
	cardinality     api.RelationCardinality
	limit           int
	depth           int
	ordinal         int
	parentOrdinal   int
	siblingOrdinal  int
	binding         includePlanBinding
	stateSeal       [sha256.Size]byte
	validated       bool
}

func (node IncludeNode) Validate() error {
	if !node.validated || node.sourceCatalog.Validate() != nil ||
		node.targetCatalog.Validate() != nil || node.targetStorage.Validate() != nil ||
		node.relationStorage.Validate() != nil || node.sourceStable.Validate() != nil ||
		node.targetStable.Validate() != nil || node.path.Validate() != nil ||
		strings.TrimSpace(node.relationName) != node.relationName ||
		node.relationName == "" || node.cardinality.Validate() != nil ||
		node.limit < 1 || node.limit > 100 || node.depth < 1 || node.depth > 2 ||
		node.ordinal < 0 || node.ordinal > 2 || node.siblingOrdinal < 0 ||
		node.siblingOrdinal > 2 || node.binding.validate() != nil ||
		node.stateSeal == [sha256.Size]byte{} || !validIncludeNodeParent(node) ||
		!validIncludeStableField(node.sourceStable) ||
		!validIncludeStableField(node.targetStable) {
		return ErrContract
	}
	if node.cardinality == api.RelationCardinalityToOne && node.limit != 1 {
		return ErrContract
	}
	seal, err := includeNodeSeal(node)
	if err != nil || seal != node.stateSeal {
		return ErrContract
	}
	return nil
}

func validIncludeNodeParent(node IncludeNode) bool {
	if node.depth == 1 {
		return node.parentOrdinal == -1
	}
	return node.parentOrdinal >= 0 && node.parentOrdinal < node.ordinal
}

func validIncludeStableField(field api.FieldDefinition) bool {
	fieldType, err := field.Type()
	if err != nil {
		return false
	}
	kind, err := fieldType.Kind()
	return err == nil && kind == api.LogicalTypeUUID
}

func (IncludeNode) String() string   { return compilerRedaction }
func (IncludeNode) GoString() string { return compilerRedaction }
func (IncludeNode) LogValue() slog.Value {
	return slog.StringValue(compilerRedaction)
}
func (node IncludeNode) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (IncludeNode) MarshalJSON() ([]byte, error)        { return rejectJSON() }
