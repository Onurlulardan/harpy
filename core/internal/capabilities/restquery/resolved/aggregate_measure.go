package resolved

import (
	"fmt"
	"math"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type AggregateMeasure struct {
	key       OutputKey
	function  api.AggregateFunction
	field     api.FieldDefinition
	hasField  bool
	path      RequestPath
	validated bool
}

func NewAggregateMeasure(
	key OutputKey,
	function api.AggregateFunction,
	field *api.FieldDefinition,
	path RequestPath,
) (AggregateMeasure, error) {
	measure := AggregateMeasure{
		key: key, function: function, path: path, validated: true,
	}
	if field != nil {
		measure.field, measure.hasField = *field, true
	}
	if err := measure.Validate(); err != nil {
		return AggregateMeasure{}, err
	}
	return measure, nil
}

func (measure AggregateMeasure) Validate() error {
	if !measure.validated || measure.key.Validate() != nil || measure.path.Validate() != nil ||
		measure.function.String() == "" || measure.hasField != (measure.field.Validate() == nil) {
		return fmt.Errorf("resolved REST Query aggregate measure is invalid")
	}
	key, err := measure.key.Value()
	if err != nil || !validMeasureOutputKey(key) {
		return fmt.Errorf("resolved REST Query aggregate measure is invalid")
	}
	if measure.function == api.AggregateFunctionCount {
		if measure.hasField {
			return fmt.Errorf("resolved REST Query aggregate measure is invalid")
		}
	} else if !measure.hasField || !measure.fieldAllowsFunction() {
		return fmt.Errorf("resolved REST Query aggregate measure is invalid")
	}
	_, resultErr := measure.resultType()
	return resultErr
}

func validMeasureOutputKey(value string) bool {
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

func (measure AggregateMeasure) fieldAllowsFunction() bool {
	capabilities, err := measure.field.Capabilities()
	if err != nil {
		return false
	}
	allowed, err := capabilities.AllowsAggregate(measure.function)
	return err == nil && allowed
}

func (measure AggregateMeasure) Key() (OutputKey, error) {
	if err := measure.Validate(); err != nil {
		return OutputKey{}, err
	}
	return measure.key, nil
}

func (measure AggregateMeasure) Function() (api.AggregateFunction, error) {
	if err := measure.Validate(); err != nil {
		return 0, err
	}
	return measure.function, nil
}

func (measure AggregateMeasure) Field() (api.FieldDefinition, bool, error) {
	if err := measure.Validate(); err != nil {
		return api.FieldDefinition{}, false, err
	}
	return measure.field, measure.hasField, nil
}

func (measure AggregateMeasure) Path() (RequestPath, error) {
	if err := measure.Validate(); err != nil {
		return RequestPath{}, err
	}
	return measure.path, nil
}

func (measure AggregateMeasure) ResultType() (api.FieldType, error) {
	if err := measure.Validate(); err != nil {
		return api.FieldType{}, err
	}
	return measure.resultType()
}

func (measure AggregateMeasure) resultType() (api.FieldType, error) {
	if measure.function == api.AggregateFunctionCount ||
		measure.function == api.AggregateFunctionCountDistinct {
		profile, err := api.NewIntegerProfile(0, math.MaxInt64)
		if err != nil {
			return api.FieldType{}, err
		}
		return api.NewIntegerFieldType(profile)
	}
	fieldType, err := measure.field.Type()
	if err != nil {
		return api.FieldType{}, err
	}
	if measure.function == api.AggregateFunctionMin ||
		measure.function == api.AggregateFunctionMax {
		return fieldType, nil
	}
	return numericAggregateType(fieldType, measure.function)
}

func (AggregateMeasure) String() string                         { return valueRedaction }
func (AggregateMeasure) GoString() string                       { return valueRedaction }
func (measure AggregateMeasure) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (AggregateMeasure) MarshalJSON() ([]byte, error)           { return rejectDirectJSONMarshal() }
