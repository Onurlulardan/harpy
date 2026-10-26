package api

import (
	"errors"
	"fmt"
)

var errInvalidFilterCapability = errors.New("REST Query filter capability is invalid")

type FilterCapabilityInput struct {
	Operator  Operator
	CountSafe bool
}

type FilterCapability struct {
	operator  Operator
	countSafe bool
	validated bool
}

func newFilterCapability(input FilterCapabilityInput) (FilterCapability, error) {
	capability := FilterCapability{
		operator: input.Operator, countSafe: input.CountSafe, validated: true,
	}
	if err := capability.Validate(); err != nil {
		return FilterCapability{}, err
	}
	return capability, nil
}

func (capability FilterCapability) Validate() error {
	if _, valid := capability.operator.Arity(); !capability.validated || !valid {
		return errInvalidFilterCapability
	}
	return nil
}

func (capability FilterCapability) Operator() (Operator, error) {
	if err := capability.Validate(); err != nil {
		return 0, err
	}
	return capability.operator, nil
}

func (capability FilterCapability) CountSafe() (bool, error) {
	if err := capability.Validate(); err != nil {
		return false, err
	}
	return capability.countSafe, nil
}

func (FilterCapability) String() string   { return queryValueRedaction }
func (FilterCapability) GoString() string { return queryValueRedaction }
func (FilterCapability) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (FilterCapability) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func (FilterCapabilityInput) String() string   { return queryValueRedaction }
func (FilterCapabilityInput) GoString() string { return queryValueRedaction }
func (FilterCapabilityInput) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (FilterCapabilityInput) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}
