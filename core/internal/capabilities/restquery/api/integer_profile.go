package api

import "fmt"

type IntegerProfile struct {
	minimum   int64
	maximum   int64
	validated bool
}

func NewIntegerProfile(minimum, maximum int64) (IntegerProfile, error) {
	profile := IntegerProfile{minimum: minimum, maximum: maximum, validated: true}
	if err := profile.Validate(); err != nil {
		return IntegerProfile{}, err
	}
	return profile, nil
}

func (profile IntegerProfile) Minimum() (int64, error) {
	if err := profile.Validate(); err != nil {
		return 0, err
	}
	return profile.minimum, nil
}

func (profile IntegerProfile) Maximum() (int64, error) {
	if err := profile.Validate(); err != nil {
		return 0, err
	}
	return profile.maximum, nil
}

func (profile IntegerProfile) Validate() error {
	if !profile.validated || profile.minimum > profile.maximum {
		return fmt.Errorf("REST Query integer profile is invalid")
	}
	return nil
}

func (profile IntegerProfile) String() string   { return queryValueRedaction }
func (profile IntegerProfile) GoString() string { return queryValueRedaction }
func (profile IntegerProfile) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (IntegerProfile) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
