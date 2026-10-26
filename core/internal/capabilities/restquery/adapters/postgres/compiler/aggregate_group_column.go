package compiler

import (
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

type AggregateGroupColumn struct {
	fieldID   uuid.UUID
	fieldType api.FieldType
	outputKey string
	validated bool
}

func newAggregateGroupColumn(
	field api.FieldDefinition, outputKey string,
) (AggregateGroupColumn, error) {
	fieldID, idErr := field.ID()
	fieldType, typeErr := field.Type()
	cloned, cloneErr := fieldType.Clone()
	if idErr != nil || typeErr != nil || cloneErr != nil {
		return AggregateGroupColumn{}, ErrContract
	}
	column := AggregateGroupColumn{
		fieldID: fieldID, fieldType: cloned,
		outputKey: outputKey, validated: true,
	}
	if column.Validate() != nil {
		return AggregateGroupColumn{}, ErrContract
	}
	return column, nil
}

func (column AggregateGroupColumn) Validate() error {
	if !column.validated || column.fieldID == uuid.Nil ||
		column.fieldID.Variant() != uuid.RFC4122 || column.fieldType.Validate() != nil ||
		!validAggregateOutputKey(column.outputKey, false) {
		return ErrContract
	}
	return nil
}

func (column AggregateGroupColumn) FieldID() (uuid.UUID, error) {
	if column.Validate() != nil {
		return uuid.Nil, ErrContract
	}
	return column.fieldID, nil
}

func (column AggregateGroupColumn) FieldType() (api.FieldType, error) {
	if column.Validate() != nil {
		return api.FieldType{}, ErrContract
	}
	return column.fieldType.Clone()
}

func (column AggregateGroupColumn) OutputKey() (string, error) {
	if column.Validate() != nil {
		return "", ErrContract
	}
	return column.outputKey, nil
}

func (AggregateGroupColumn) String() string   { return compilerRedaction }
func (AggregateGroupColumn) GoString() string { return compilerRedaction }
func (AggregateGroupColumn) LogValue() slog.Value {
	return slog.StringValue(compilerRedaction)
}
func (column AggregateGroupColumn) Format(state fmt.State, _ rune) {
	writeRedaction(state)
}
func (AggregateGroupColumn) MarshalJSON() ([]byte, error) { return rejectJSON() }
