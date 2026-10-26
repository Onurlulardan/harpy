package api

import "fmt"

const maximumDecimalPrecision = 1000

type DecimalProfile struct {
	precision int
	scale     int
	validated bool
}

func NewDecimalProfile(precision, scale int) (DecimalProfile, error) {
	profile := DecimalProfile{precision: precision, scale: scale, validated: true}
	if err := profile.Validate(); err != nil {
		return DecimalProfile{}, err
	}
	return profile, nil
}

func (profile DecimalProfile) Precision() (int, error) {
	if err := profile.Validate(); err != nil {
		return 0, err
	}
	return profile.precision, nil
}

func (profile DecimalProfile) Scale() (int, error) {
	if err := profile.Validate(); err != nil {
		return 0, err
	}
	return profile.scale, nil
}

func (profile DecimalProfile) Validate() error {
	if !profile.validated || profile.precision < 1 ||
		profile.precision > maximumDecimalPrecision ||
		profile.scale < 0 || profile.scale > profile.precision {
		return fmt.Errorf("REST Query decimal profile is invalid")
	}
	return nil
}

func (profile DecimalProfile) String() string   { return queryValueRedaction }
func (profile DecimalProfile) GoString() string { return queryValueRedaction }
func (profile DecimalProfile) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (DecimalProfile) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
