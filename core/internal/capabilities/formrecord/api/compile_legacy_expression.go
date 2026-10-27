package recordkernel

import (
	"github.com/google/uuid"
)

func compileLegacyExpression(
	source Expression,
	expected TypeID,
	fields map[uuid.UUID]Field,
	depth uint64,
	limits BudgetLimits,
) (expression, uint64, map[uuid.UUID]struct{}, error) {
	if expected == "" || depth > limits.CalculationASTDepth {
		return expression{}, 0, nil, ErrCatalogInvalid
	}
	result := expression{kind: source.Kind, typeID: expected}
	refs := map[uuid.UUID]struct{}{}
	if source.Kind == "literal" {
		value, err := legacyLiteral(expected, source.Literal)
		if err != nil || source.FieldID != nil || source.Operator != nil ||
			source.Function != nil || source.Arguments != nil {
			return expression{}, 0, nil, ErrCatalogInvalid
		}
		result.literal = value
		return result, 1, refs, nil
	}
	if source.Kind == "field" {
		return compileLegacyField(source, expected, fields)
	}
	childTypes, err := legacyChildTypes(source, expected, fields)
	if err != nil || len(childTypes) != len(source.Arguments) ||
		len(childTypes) > MaximumExpressionArguments {
		return expression{}, 0, nil, ErrCatalogInvalid
	}
	if source.Operator != nil {
		result.operator = *source.Operator
	}
	if source.Function != nil {
		result.function = *source.Function
	}
	count := uint64(1)
	result.arguments = make([]expression, len(source.Arguments))
	for index, argument := range source.Arguments {
		compiled, nodes, childRefs, childErr := compileLegacyExpression(
			argument, childTypes[index], fields, depth+1, limits,
		)
		if childErr != nil {
			return expression{}, 0, nil, childErr
		}
		result.arguments[index] = compiled
		count += nodes
		if count > limits.CalculationASTNodes {
			return expression{}, 0, nil, ErrCatalogInvalid
		}
		for fieldID := range childRefs {
			refs[fieldID] = struct{}{}
		}
	}
	if !validLegacyShape(source) || !validExpressionSignature(result) {
		return expression{}, 0, nil, ErrCatalogInvalid
	}
	return result, count, refs, nil
}

func compileLegacyField(
	source Expression,
	expected TypeID,
	fields map[uuid.UUID]Field,
) (expression, uint64, map[uuid.UUID]struct{}, error) {
	if source.FieldID == nil || source.Literal != nil || source.Operator != nil ||
		source.Function != nil || source.Arguments != nil {
		return expression{}, 0, nil, ErrCatalogInvalid
	}
	fieldID, err := parseStableID(*source.FieldID)
	field, found := fields[fieldID]
	_, supported := expressionType(string(field.TypeID))
	if err != nil || !found || field.TypeID != expected || !supported {
		return expression{}, 0, nil, ErrCatalogInvalid
	}
	return expression{kind: "field", typeID: expected, fieldID: fieldID},
		1, map[uuid.UUID]struct{}{fieldID: {}}, nil
}
