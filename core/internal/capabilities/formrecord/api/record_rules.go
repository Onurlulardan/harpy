package recordkernel

import "encoding/json"

type RecordRules struct {
	Fields                []RecordFieldRule      `json:"fields"`
	WorkflowActions       []string               `json:"workflow_actions"`
	Defaults              []DefaultRule          `json:"defaults"`
	CrossFieldValidations []CrossFieldValidation `json:"cross_field_validations"`
}

type RecordFieldRule struct {
	FieldID     string      `json:"field_id"`
	WritePolicy WritePolicy `json:"write_policy"`
}

type DefaultRule struct {
	FieldID    string          `json:"field_id"`
	Expression TypedExpression `json:"expression"`
}

type CrossFieldValidation struct {
	ID        string          `json:"id"`
	Code      string          `json:"code"`
	Predicate TypedExpression `json:"predicate"`
}

type TypedExpression struct {
	Kind      string            `json:"kind"`
	Type      string            `json:"type"`
	FieldID   *string           `json:"field_id,omitempty"`
	Value     json.RawMessage   `json:"value,omitempty"`
	Operator  *string           `json:"operator,omitempty"`
	Function  *string           `json:"function,omitempty"`
	Arguments []TypedExpression `json:"arguments,omitempty"`
}
