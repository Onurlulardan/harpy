package recordkernel

import "encoding/json"

type catalogExpression struct {
	Kind      string              `json:"kind"`
	Type      string              `json:"type"`
	FieldID   string              `json:"field_id,omitempty"`
	Value     json.RawMessage     `json:"value,omitempty"`
	Operator  string              `json:"operator,omitempty"`
	Function  string              `json:"function,omitempty"`
	Arguments []catalogExpression `json:"arguments,omitempty"`
}

type catalogAssignment struct {
	FieldID    string            `json:"field_id"`
	Expression catalogExpression `json:"expression"`
}

type catalogCrossRule struct {
	ID        string            `json:"id"`
	Code      string            `json:"code"`
	Predicate catalogExpression `json:"predicate"`
}

func expressionDocument(value expression) (catalogExpression, error) {
	document := catalogExpression{Kind: value.kind, Type: string(value.typeID)}
	switch value.kind {
	case "literal":
		document.Value = value.literal.CanonicalJSON()
		if len(document.Value) == 0 {
			return catalogExpression{}, ErrCatalogInvalid
		}
	case "field":
		document.FieldID = value.fieldID.String()
	case "unary", "binary":
		document.Operator = value.operator
	case "call":
		document.Function = value.function
	default:
		return catalogExpression{}, ErrCatalogInvalid
	}
	if len(value.arguments) > 0 {
		document.Arguments = make([]catalogExpression, len(value.arguments))
		for index, argument := range value.arguments {
			encoded, err := expressionDocument(argument)
			if err != nil {
				return catalogExpression{}, err
			}
			document.Arguments[index] = encoded
		}
	}
	return document, nil
}
