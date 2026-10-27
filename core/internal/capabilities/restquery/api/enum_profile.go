package api

import (
	"fmt"

	"github.com/google/uuid"
)

type EnumProfile struct {
	version   uuid.UUID
	values    []EnumValue
	ordered   bool
	validated bool
}

func NewEnumProfile(version string, values []EnumValue) (EnumProfile, error) {
	parsed, err := uuid.Parse(version)
	if err != nil || parsed == uuid.Nil || parsed.String() != version ||
		parsed.Variant() != uuid.RFC4122 {
		return EnumProfile{}, fmt.Errorf("REST Query enum profile is invalid")
	}
	profile := EnumProfile{
		version: parsed, values: append([]EnumValue{}, values...), validated: true,
	}
	if len(profile.values) > 0 {
		profile.ordered = profile.values[0].hasOrdinal
	}
	if err := profile.Validate(); err != nil {
		return EnumProfile{}, err
	}
	return profile, nil
}

func (profile EnumProfile) Version() (uuid.UUID, error) {
	if err := profile.Validate(); err != nil {
		return uuid.Nil, err
	}
	return profile.version, nil
}

func (profile EnumProfile) Ordered() (bool, error) {
	if err := profile.Validate(); err != nil {
		return false, err
	}
	return profile.ordered, nil
}

func (profile EnumProfile) Values() ([]EnumValue, error) {
	if err := profile.Validate(); err != nil {
		return nil, err
	}
	return append([]EnumValue{}, profile.values...), nil
}

func (profile EnumProfile) Lookup(code string) (EnumValue, bool, error) {
	if err := profile.Validate(); err != nil {
		return EnumValue{}, false, err
	}
	for _, value := range profile.values {
		if value.code == code {
			return value, true, nil
		}
	}
	return EnumValue{}, false, nil
}

func (profile EnumProfile) LookupOrdinal(ordinal int) (EnumValue, bool, error) {
	if err := profile.Validate(); err != nil {
		return EnumValue{}, false, err
	}
	if !profile.ordered {
		return EnumValue{}, false, nil
	}
	for _, value := range profile.values {
		if value.ordinal == ordinal {
			return value, true, nil
		}
	}
	return EnumValue{}, false, nil
}

func (profile EnumProfile) Validate() error {
	if !profile.validated || profile.version == uuid.Nil || len(profile.values) == 0 {
		return fmt.Errorf("REST Query enum profile is invalid")
	}
	codes := make(map[string]struct{}, len(profile.values))
	ordinals := make(map[int]struct{}, len(profile.values))
	for _, value := range profile.values {
		if value.Validate() != nil || value.hasOrdinal != profile.ordered {
			return fmt.Errorf("REST Query enum profile is invalid")
		}
		if _, duplicate := codes[value.code]; duplicate {
			return fmt.Errorf("REST Query enum profile is invalid")
		}
		codes[value.code] = struct{}{}
		if profile.ordered {
			if _, duplicate := ordinals[value.ordinal]; duplicate {
				return fmt.Errorf("REST Query enum profile is invalid")
			}
			ordinals[value.ordinal] = struct{}{}
		}
	}
	return nil
}

func (profile EnumProfile) String() string   { return queryValueRedaction }
func (profile EnumProfile) GoString() string { return queryValueRedaction }
func (profile EnumProfile) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (EnumProfile) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func (profile EnumProfile) clone() EnumProfile {
	result := profile
	result.values = append([]EnumValue{}, profile.values...)
	return result
}
