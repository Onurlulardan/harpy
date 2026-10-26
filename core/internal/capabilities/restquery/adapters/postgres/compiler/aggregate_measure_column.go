package compiler

import (
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

type AggregateMeasureColumn struct {
	sourceID  uuid.UUID
	fieldType api.FieldType
	outputKey string
	function  api.AggregateFunction
	hasSource bool
	validated bool
}

func newAggregateMeasureColumn(
	measure aggregateMeasure,
) (AggregateMeasureColumn, error) {
	resultType, err := measure.declaration.ResultType()
	if err != nil {
		return AggregateMeasureColumn{}, ErrContract
	}
	cloned, err := resultType.Clone()
	if err != nil {
		return AggregateMeasureColumn{}, ErrContract
	}
	column := AggregateMeasureColumn{
		fieldType: cloned, outputKey: measure.key, function: measure.function,
		hasSource: measure.hasField, validated: true,
	}
	if measure.hasField {
		column.sourceID, err = measure.field.ID()
		if err != nil {
			return AggregateMeasureColumn{}, ErrContract
		}
	}
	if column.Validate() != nil {
		return AggregateMeasureColumn{}, ErrContract
	}
	return column, nil
}

func (column AggregateMeasureColumn) Validate() error {
	if !column.validated || column.fieldType.Validate() != nil ||
		column.function.String() == "" ||
		!validAggregateOutputKey(column.outputKey, true) ||
		column.hasSource != (column.sourceID != uuid.Nil) ||
		column.hasSource && column.sourceID.Variant() != uuid.RFC4122 ||
		(column.function == api.AggregateFunctionCount) == column.hasSource {
		return ErrContract
	}
	return nil
}

func (column AggregateMeasureColumn) FieldType() (api.FieldType, error) {
	if column.Validate() != nil {
		return api.FieldType{}, ErrContract
	}
	return column.fieldType.Clone()
}

func (column AggregateMeasureColumn) OutputKey() (string, error) {
	if column.Validate() != nil {
		return "", ErrContract
	}
	return column.outputKey, nil
}

func (column AggregateMeasureColumn) Function() (api.AggregateFunction, error) {
	if column.Validate() != nil {
		return 0, ErrContract
	}
	return column.function, nil
}

func (AggregateMeasureColumn) String() string   { return compilerRedaction }
func (AggregateMeasureColumn) GoString() string { return compilerRedaction }
func (AggregateMeasureColumn) LogValue() slog.Value {
	return slog.StringValue(compilerRedaction)
}
func (column AggregateMeasureColumn) Format(state fmt.State, _ rune) {
	writeRedaction(state)
}
func (AggregateMeasureColumn) MarshalJSON() ([]byte, error) { return rejectJSON() }
