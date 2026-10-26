package executor

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/recordpage"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func (graph materialGraph) renderRecord(
	nodeIndex int,
	record materialRecord,
	occurrences *int,
) ([]byte, error) {
	if occurrences == nil || record.validate() != nil {
		return nil, ErrContract
	}
	*occurrences = *occurrences + 1
	if *occurrences > maximumIncludedOccurrences {
		return nil, includeResultBudgetError()
	}
	included, err := graph.renderIncluded(nodeIndex, record.id, occurrences)
	if err != nil {
		return nil, err
	}
	document, err := recordpage.AppendIncluded(record.document, included)
	if err != nil {
		return nil, ErrContract
	}
	return document, nil
}

func (node materialNode) selectRecords(
	source uuid.UUID,
) ([]materialRecord, error) {
	targets := node.associations[source]
	if len(targets) == 0 {
		return []materialRecord{}, nil
	}
	result := make([]materialRecord, 0, node.limit)
	visible := 0
	for _, record := range node.records {
		if _, associated := targets[record.id]; !associated {
			continue
		}
		visible++
		if node.cardinality == api.RelationCardinalityToOne {
			if visible > 1 {
				return nil, relationCardinalityError()
			}
			result = append(result, record)
			continue
		}
		if len(result) < node.limit {
			result = append(result, record)
		}
	}
	return result, nil
}
