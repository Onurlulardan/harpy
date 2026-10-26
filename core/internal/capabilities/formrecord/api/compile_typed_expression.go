package recordkernel

import (
	"github.com/google/uuid"
)

func compileTypedExpression(
	source TypedExpression,
	expected TypeID,
	fields map[uuid.UUID]Field,
	depth uint64,
	limits BudgetLimits,
) (expression, uint64, map[uuid.UUID]struct{}, error) {
	typeID, known := expressionType(source.Type)
	if !known || expected != "" && typeID != expected ||
		depth > limits.CalculationASTDepth {
		return expression{}, 0, nil, ErrCatalogInvalid
	}
	result := expression{kind: source.Kind, typeID: typeID}
	refs := map[uuid.UUID]struct{}{}
	if err := compileTypedLeaf(source, fields, &result, refs); err != nil {
		return expression{}, 0, nil, err
	}
	count := uint64(1)
	if len(source.Arguments) > MaximumExpressionArguments {
		return expression{}, 0, nil, ErrCatalogInvalid
	}
	result.arguments = make([]expression, len(source.Arguments))
	for index, argument := range source.Arguments {
		compiled, nodes, childRefs, err := compileTypedExpression(
			argument, "", fields, depth+1, limits,
		)
		if err != nil {
			return expression{}, 0, nil, err
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
	if !validExpressionSignature(result) {
		return expression{}, 0, nil, ErrCatalogInvalid
	}
	return result, count, refs, nil
}

func compileTypedLeaf(
	source TypedExpression,
	fields map[uuid.UUID]Field,
	result *expression,
	refs map[uuid.UUID]struct{},
) error {
	switch source.Kind {
	case "literal":
		if source.FieldID != nil || source.Operator != nil || source.Function != nil ||
			source.Arguments != nil || len(source.Value) == 0 {
			return ErrCatalogInvalid
		}
		value, err := decodeExpressionLiteral(result.typeID, source.Value)
		result.literal = value
		return err
	case "field":
		if source.FieldID == nil || source.Operator != nil || source.Function != nil ||
			source.Arguments != nil || source.Value != nil {
			return ErrCatalogInvalid
		}
		fieldID, err := parseStableID(*source.FieldID)
		field, found := fields[fieldID]
		if err != nil || !found || field.TypeID != result.typeID {
			return ErrCatalogInvalid
		}
		result.fieldID, refs[fieldID] = fieldID, struct{}{}
		return nil
	case "unary":
		if source.Operator == nil || source.FieldID != nil || source.Function != nil ||
			source.Value != nil {
			return ErrCatalogInvalid
		}
		result.operator = *source.Operator
	case "binary":
		if source.Operator == nil || source.FieldID != nil || source.Function != nil ||
			source.Value != nil {
			return ErrCatalogInvalid
		}
		result.operator = *source.Operator
	case "call":
		if source.Function == nil || source.FieldID != nil || source.Operator != nil ||
			source.Value != nil {
			return ErrCatalogInvalid
		}
		result.function = *source.Function
	default:
		return ErrCatalogInvalid
	}
	return nil
}
