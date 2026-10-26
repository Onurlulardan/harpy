package api

import "fmt"

const (
	TextExactCollationPGUnicodeFast = "pg_unicode_fast"
	TextExactNormalizationNFC       = "NFC"
)

type TextCasefoldProfile uint8

const (
	TextCasefoldDisabled TextCasefoldProfile = iota + 1
	TextCasefoldUnicodeDefaultV1
)

type TextAccentProfile uint8

const (
	TextAccentDisabled TextAccentProfile = iota + 1
	TextAccentInsensitiveV1
)

type TextProfile struct {
	casefold  TextCasefoldProfile
	accent    TextAccentProfile
	validated bool
}

func NewTextProfile(
	casefold TextCasefoldProfile, accent TextAccentProfile,
) (TextProfile, error) {
	profile := TextProfile{casefold: casefold, accent: accent, validated: true}
	if err := profile.Validate(); err != nil {
		return TextProfile{}, err
	}
	return profile, nil
}

func (profile TextProfile) ExactCollation() (string, error) {
	if err := profile.Validate(); err != nil {
		return "", err
	}
	return TextExactCollationPGUnicodeFast, nil
}
func (profile TextProfile) ExactNormalization() (string, error) {
	if err := profile.Validate(); err != nil {
		return "", err
	}
	return TextExactNormalizationNFC, nil
}
func (profile TextProfile) Casefold() (TextCasefoldProfile, bool, error) {
	if err := profile.Validate(); err != nil {
		return 0, false, err
	}
	return profile.casefold, profile.casefold != TextCasefoldDisabled, nil
}
func (profile TextProfile) Accent() (TextAccentProfile, bool, error) {
	if err := profile.Validate(); err != nil {
		return 0, false, err
	}
	return profile.accent, profile.accent != TextAccentDisabled, nil
}

func (profile TextProfile) Validate() error {
	if !profile.validated || !profile.casefold.valid() || !profile.accent.valid() ||
		profile.accent == TextAccentInsensitiveV1 &&
			profile.casefold != TextCasefoldUnicodeDefaultV1 {
		return fmt.Errorf("REST Query text profile is invalid")
	}
	return nil
}

func (profile TextProfile) String() string   { return queryValueRedaction }
func (profile TextProfile) GoString() string { return queryValueRedaction }
func (profile TextProfile) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (TextProfile) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func (profile TextCasefoldProfile) String() string {
	switch profile {
	case TextCasefoldDisabled:
		return "none"
	case TextCasefoldUnicodeDefaultV1:
		return "unicode_default_casefold_v1"
	default:
		return ""
	}
}

func (profile TextCasefoldProfile) valid() bool {
	return profile == TextCasefoldDisabled || profile == TextCasefoldUnicodeDefaultV1
}

func (profile TextAccentProfile) String() string {
	switch profile {
	case TextAccentDisabled:
		return "none"
	case TextAccentInsensitiveV1:
		return "accent_insensitive_v1"
	default:
		return ""
	}
}

func (profile TextAccentProfile) valid() bool {
	return profile == TextAccentDisabled || profile == TextAccentInsensitiveV1
}

func (TextCasefoldProfile) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
func (TextAccentProfile) MarshalJSON() ([]byte, error)   { return rejectDirectJSONMarshal() }
