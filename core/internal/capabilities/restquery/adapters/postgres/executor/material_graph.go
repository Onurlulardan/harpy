package executor

import (
	"strings"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

const maximumIncludedOccurrences = 1000

type materialNode struct {
	relation     string
	cardinality  api.RelationCardinality
	limit        int
	parent       int
	depth        int
	ordinal      int
	records      []materialRecord
	associations map[uuid.UUID]map[uuid.UUID]struct{}
}

type materialGraph struct {
	root  []materialRecord
	nodes []materialNode
}

func (graph materialGraph) validate() error {
	if graph.root == nil || len(graph.root) > 500 ||
		len(graph.nodes) < 1 || len(graph.nodes) > 3 ||
		!uniqueMaterialRecords(graph.root) {
		return ErrContract
	}
	siblings := make(map[int]map[string]struct{}, len(graph.nodes))
	for index, node := range graph.nodes {
		if node.validate(index) != nil {
			return ErrContract
		}
		if node.depth == 2 && graph.nodes[node.parent].depth != 1 {
			return ErrContract
		}
		if siblings[node.parent] == nil {
			siblings[node.parent] = make(map[string]struct{})
		}
		if _, duplicate := siblings[node.parent][node.relation]; duplicate {
			return ErrContract
		}
		siblings[node.parent][node.relation] = struct{}{}
	}
	return nil
}

func (node materialNode) validate(index int) error {
	if !validRelationName(node.relation) || node.cardinality.Validate() != nil ||
		node.limit < 1 || node.limit > 100 || node.depth < 1 || node.depth > 2 ||
		node.ordinal != index || node.parent < -1 || node.parent >= index ||
		node.depth == 1 && node.parent != -1 ||
		node.depth == 2 && (node.parent < 0 || node.parent >= index) ||
		node.records == nil || node.associations == nil ||
		!uniqueMaterialRecords(node.records) {
		return ErrContract
	}
	if node.cardinality == api.RelationCardinalityToOne && node.limit != 1 {
		return ErrContract
	}
	for source, targets := range node.associations {
		if source == uuid.Nil || source.Variant() != uuid.RFC4122 ||
			targets == nil || len(targets) > 100 {
			return ErrContract
		}
		for target := range targets {
			if target == uuid.Nil || target.Variant() != uuid.RFC4122 {
				return ErrContract
			}
		}
	}
	return nil
}

func uniqueMaterialRecords(records []materialRecord) bool {
	seen := make(map[uuid.UUID]struct{}, len(records))
	for _, record := range records {
		if record.validate() != nil {
			return false
		}
		if _, duplicate := seen[record.id]; duplicate {
			return false
		}
		seen[record.id] = struct{}{}
	}
	return true
}

func validRelationName(value string) bool {
	return value != "" && value != "_included" && utf8.ValidString(value) &&
		!strings.ContainsRune(value, '\x00') && norm.NFC.IsNormalString(value)
}
