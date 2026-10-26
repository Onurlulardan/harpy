package published

import "fmt"

type ParameterArity uint8

const (
	ParameterAritySingle ParameterArity = iota + 1
	ParameterArityList
)

func (arity ParameterArity) String() string {
	switch arity {
	case ParameterAritySingle:
		return "single"
	case ParameterArityList:
		return "list"
	default:
		return ""
	}
}

func (arity ParameterArity) Valid() bool {
	return arity == ParameterAritySingle || arity == ParameterArityList
}

func (arity ParameterArity) Validate() error {
	if !arity.Valid() {
		return fmt.Errorf("published parameter arity is invalid")
	}
	return nil
}

func (ParameterArity) MarshalJSON() ([]byte, error) { return rejectPublishedJSON() }
