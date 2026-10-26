package api

import "fmt"

const maximumAggregateMeasures = 8

type RequestMeasure struct {
	name     string
	function AggregateFunction
	field    string
	hasField bool
}

func NewRequestMeasure(
	name string, function AggregateFunction, field *string,
) (RequestMeasure, error) {
	measure := RequestMeasure{name: name, function: function}
	if field != nil {
		measure.field, measure.hasField = *field, true
	}
	if !measure.valid() {
		return RequestMeasure{}, fmt.Errorf("REST Query aggregate measure is invalid")
	}
	return measure, nil
}

func (measure RequestMeasure) Name() string                { return measure.name }
func (measure RequestMeasure) Function() AggregateFunction { return measure.function }
func (measure RequestMeasure) Field() (string, bool)       { return measure.field, measure.hasField }
func (measure RequestMeasure) String() string              { return queryValueRedaction }
func (measure RequestMeasure) GoString() string            { return queryValueRedaction }
func (measure RequestMeasure) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (RequestMeasure) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func (measure RequestMeasure) valid() bool {
	if !validMeasureName(measure.name) || !measure.function.valid() {
		return false
	}
	if measure.function == AggregateFunctionCount {
		return !measure.hasField && measure.field == ""
	}
	return measure.hasField && validRequestName(measure.field)
}

func validMeasureName(value string) bool {
	if len(value) < 1 || len(value) > 63 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}

func cloneRequestMeasures(values []RequestMeasure) ([]RequestMeasure, error) {
	if len(values) > maximumAggregateMeasures {
		return nil, fmt.Errorf("REST Query aggregate measure count exceeds its limit")
	}
	result := append([]RequestMeasure{}, values...)
	seen := make(map[string]struct{}, len(result))
	for index, value := range result {
		if !value.valid() {
			return nil, fmt.Errorf("REST Query aggregate measure %d is invalid", index)
		}
		if _, duplicate := seen[value.name]; duplicate {
			return nil, fmt.Errorf("REST Query aggregate measure name is duplicated")
		}
		seen[value.name] = struct{}{}
	}
	return result, nil
}
