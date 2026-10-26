package recordkernel

import (
	"github.com/google/uuid"
)

func decodeCrossRules(
	documents []catalogCrossRule,
	fields map[uuid.UUID]Field,
	limits BudgetLimits,
) ([]crossFieldRule, uint64, error) {
	if documents == nil || uint64(len(documents)) > limits.ValidationRules {
		return nil, 0, ErrCatalogInvalid
	}
	result := make([]crossFieldRule, len(documents))
	seenCodes := make(map[string]struct{}, len(documents))
	var total uint64
	for index, document := range documents {
		ruleID, err := parseCatalogID(document.ID)
		if err != nil || !ruleCodePattern.MatchString(document.Code) ||
			index > 0 && document.ID <= documents[index-1].ID {
			return nil, 0, ErrCatalogInvalid
		}
		if _, duplicate := seenCodes[document.Code]; duplicate {
			return nil, 0, ErrCatalogInvalid
		}
		predicate, nodes, _, err := decodeCatalogExpression(
			document.Predicate, TypeBoolean, fields, limits,
		)
		if err != nil {
			return nil, 0, err
		}
		seenCodes[document.Code] = struct{}{}
		result[index] = crossFieldRule{ruleID, document.Code, predicate}
		total += nodes
	}
	return result, total, nil
}
