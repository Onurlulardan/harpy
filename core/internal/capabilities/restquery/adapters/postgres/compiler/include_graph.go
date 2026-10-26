package compiler

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
)

type IncludeGraph struct {
	nodes     []IncludeNode
	binding   includePlanBinding
	stateSeal [sha256.Size]byte
	validated bool
}

func newIncludeGraph(
	nodes []IncludeNode, binding includePlanBinding,
) (IncludeGraph, error) {
	graph := IncludeGraph{
		nodes: append([]IncludeNode{}, nodes...), binding: binding, validated: true,
	}
	seal, err := includeGraphSeal(graph)
	if err != nil {
		return IncludeGraph{}, ErrContract
	}
	graph.stateSeal = seal
	if graph.Validate() != nil {
		return IncludeGraph{}, ErrContract
	}
	return graph, nil
}

func (graph IncludeGraph) Validate() error {
	if !graph.validated || len(graph.nodes) < 1 || len(graph.nodes) > 3 ||
		graph.binding.validate() != nil || graph.stateSeal == [sha256.Size]byte{} {
		return ErrContract
	}
	for index, node := range graph.nodes {
		if node.Validate() != nil || node.binding != graph.binding || node.ordinal != index {
			return ErrContract
		}
	}
	seal, err := includeGraphSeal(graph)
	if err != nil || seal != graph.stateSeal {
		return ErrContract
	}
	return nil
}

func (IncludeGraph) String() string   { return compilerRedaction }
func (IncludeGraph) GoString() string { return compilerRedaction }
func (IncludeGraph) LogValue() slog.Value {
	return slog.StringValue(compilerRedaction)
}
func (graph IncludeGraph) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (IncludeGraph) MarshalJSON() ([]byte, error)         { return rejectJSON() }
