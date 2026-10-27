package executor

import (
	"bytes"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/recordpage"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func materializeGraph(
	page recordpage.Page,
	graph materialGraph,
) (recordpage.Page, error) {
	if page.Validate() != nil || graph.validate() != nil {
		return recordpage.Page{}, ErrContract
	}
	records, err := page.Records()
	if err != nil || len(records) != len(graph.root) {
		return recordpage.Page{}, ErrContract
	}
	included := make([][]byte, len(graph.root))
	occurrences := 0
	for index, root := range graph.root {
		if !bytes.Equal(records[index], root.document) {
			return recordpage.Page{}, ErrContract
		}
		included[index], err = graph.renderIncluded(-1, root.id, &occurrences)
		if err != nil {
			return recordpage.Page{}, err
		}
	}
	result, err := recordpage.AttachIncluded(page, included)
	if err != nil {
		return recordpage.Page{}, ErrContract
	}
	return result, nil
}

func (graph materialGraph) renderIncluded(
	parent int,
	source uuid.UUID,
	occurrences *int,
) ([]byte, error) {
	var document bytes.Buffer
	document.WriteByte('{')
	written := 0
	for index := range graph.nodes {
		node := graph.nodes[index]
		if node.parent != parent {
			continue
		}
		value, err := graph.renderRelation(index, source, occurrences)
		if err != nil {
			return nil, err
		}
		key, _ := json.Marshal(node.relation)
		if written > 0 {
			document.WriteByte(',')
		}
		document.Write(key)
		document.WriteByte(':')
		document.Write(value)
		written++
	}
	document.WriteByte('}')
	return document.Bytes(), nil
}

func (graph materialGraph) renderRelation(
	nodeIndex int,
	source uuid.UUID,
	occurrences *int,
) ([]byte, error) {
	node := graph.nodes[nodeIndex]
	selected, err := node.selectRecords(source)
	if err != nil {
		return nil, err
	}
	if node.cardinality == api.RelationCardinalityToOne {
		if len(selected) == 0 {
			return []byte("null"), nil
		}
		return graph.renderRecord(nodeIndex, selected[0], occurrences)
	}
	var value bytes.Buffer
	value.WriteByte('[')
	for index, record := range selected {
		if index > 0 {
			value.WriteByte(',')
		}
		document, err := graph.renderRecord(nodeIndex, record, occurrences)
		if err != nil {
			return nil, err
		}
		value.Write(document)
	}
	value.WriteByte(']')
	return value.Bytes(), nil
}
