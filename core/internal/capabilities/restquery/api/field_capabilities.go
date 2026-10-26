package api

import (
	"errors"
	"fmt"
)

const maximumCursorCanonicalBytes = 2048

var errInvalidFieldCapabilities = errors.New("REST Query field capabilities are invalid")

type FieldCapabilitiesInput struct {
	Selectable           bool
	Filters              []FilterCapabilityInput
	Sortable             bool
	CursorCanonicalBytes int
	Searchable           bool
	Groupable            bool
	Aggregates           []AggregateFunction
}

type FieldCapabilities struct {
	logicalType          LogicalType
	orderedEnum          bool
	casefoldText         bool
	selectable           bool
	filters              []FilterCapability
	sortable             bool
	cursorCanonicalBytes int
	searchable           bool
	groupable            bool
	aggregates           []AggregateFunction
	validated            bool
}

func NewFieldCapabilities(
	fieldType FieldType, input FieldCapabilitiesInput,
) (FieldCapabilities, error) {
	kind, orderedEnum, casefoldText, err := fieldCapabilityTypeShape(fieldType)
	if err != nil {
		return FieldCapabilities{}, errInvalidFieldCapabilities
	}
	filters, err := canonicalFilterCapabilities(input.Filters)
	if err != nil {
		return FieldCapabilities{}, err
	}
	aggregates, err := canonicalAggregateCapabilities(input.Aggregates)
	if err != nil {
		return FieldCapabilities{}, err
	}
	capabilities := FieldCapabilities{
		logicalType: kind, orderedEnum: orderedEnum, casefoldText: casefoldText,
		selectable: input.Selectable, filters: filters, sortable: input.Sortable,
		cursorCanonicalBytes: input.CursorCanonicalBytes,
		searchable:           input.Searchable, groupable: input.Groupable,
		aggregates: aggregates, validated: true,
	}
	if err := capabilities.Validate(); err != nil {
		return FieldCapabilities{}, err
	}
	return capabilities, nil
}

func (capabilities FieldCapabilities) Validate() error {
	if !capabilities.validated || !capabilities.logicalType.Valid() ||
		capabilities.filters == nil || capabilities.aggregates == nil ||
		capabilities.orderedEnum && capabilities.logicalType != LogicalTypeEnum ||
		capabilities.casefoldText && capabilities.logicalType != LogicalTypeText ||
		!validCursorCapability(capabilities) ||
		capabilities.searchable && !capabilities.casefoldText ||
		capabilities.sortable &&
			!sortableLogicalType(capabilities.logicalType, capabilities.orderedEnum) ||
		capabilities.groupable && !groupableLogicalType(capabilities.logicalType) ||
		!validFilterCapabilities(capabilities) ||
		!validAggregateCapabilities(capabilities) {
		return errInvalidFieldCapabilities
	}
	return nil
}

func (FieldCapabilities) String() string   { return queryValueRedaction }
func (FieldCapabilities) GoString() string { return queryValueRedaction }
func (FieldCapabilities) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (FieldCapabilities) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}

func (FieldCapabilitiesInput) String() string   { return queryValueRedaction }
func (FieldCapabilitiesInput) GoString() string { return queryValueRedaction }
func (FieldCapabilitiesInput) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (FieldCapabilitiesInput) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}
