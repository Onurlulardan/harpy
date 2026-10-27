package recordkernel

import (
	"github.com/google/uuid"
	"regexp"
	"slices"
)

var ruleCodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{2,63}$`)

func compileCrossFieldRules(
	rules []CrossFieldValidation,
	fields map[uuid.UUID]Field,
	limits BudgetLimits,
) ([]crossFieldRule, uint64, error) {
	if uint64(len(rules)) > limits.ValidationRules {
		return nil, 0, ErrCatalogInvalid
	}
	result := make([]crossFieldRule, len(rules))
	seenIDs := make(map[uuid.UUID]struct{}, len(rules))
	seenCodes := make(map[string]struct{}, len(rules))
	var total uint64
	for index, rule := range rules {
		ruleID, err := parseStableID(rule.ID)
		if err != nil || !ruleCodePattern.MatchString(rule.Code) {
			return nil, 0, ErrCatalogInvalid
		}
		if _, duplicate := seenIDs[ruleID]; duplicate {
			return nil, 0, ErrCatalogInvalid
		}
		if _, duplicate := seenCodes[rule.Code]; duplicate {
			return nil, 0, ErrCatalogInvalid
		}
		compiled, nodes, _, err := compileTypedExpression(
			rule.Predicate, TypeBoolean, fields, 1, limits,
		)
		if err != nil {
			return nil, 0, err
		}
		seenIDs[ruleID], seenCodes[rule.Code] = struct{}{}, struct{}{}
		result[index] = crossFieldRule{ruleID, rule.Code, compiled}
		total += nodes
	}
	slices.SortFunc(result, func(left, right crossFieldRule) int {
		return compareUUID(left.id, right.id)
	})
	return result, total, nil
}
