package compiler

import "crypto/sha256"

func includeGraphSeal(graph IncludeGraph) ([sha256.Size]byte, error) {
	writer := newSealWriter("harpy.restquery.postgres.include-graph.v1")
	if writeIncludePlanBinding(writer, graph.binding) != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	writer.writeInt(int64(len(graph.nodes)))
	for _, node := range graph.nodes {
		if node.Validate() != nil {
			return [sha256.Size]byte{}, ErrContract
		}
		writer.writeBytes(node.stateSeal[:])
	}
	return writer.sum()
}
