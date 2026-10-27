package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestTextProfileFixesExactUnicodeSemantics(t *testing.T) {
	profile, err := NewTextProfile(TextCasefoldDisabled, TextAccentDisabled)
	if err != nil {
		t.Fatal(err)
	}
	collation, collationErr := profile.ExactCollation()
	normalization, normalizationErr := profile.ExactNormalization()
	if collationErr != nil || normalizationErr != nil ||
		collation != TextExactCollationPGUnicodeFast ||
		normalization != TextExactNormalizationNFC {
		t.Fatal("exact text semantics changed")
	}
	if _, found, err := profile.Casefold(); found || err != nil {
		t.Fatal("exact-only profile gained casefold")
	}
	if _, found, err := profile.Accent(); found || err != nil {
		t.Fatal("exact-only profile gained accent companion")
	}
}

func TestTextProfileAllowsOnlyVersionedCompanions(t *testing.T) {
	profile, err := NewTextProfile(
		TextCasefoldUnicodeDefaultV1, TextAccentInsensitiveV1,
	)
	if err != nil {
		t.Fatal(err)
	}
	if casefold, found, err := profile.Casefold(); err != nil || !found ||
		casefold.String() != "unicode_default_casefold_v1" {
		t.Fatal("casefold companion differs")
	}
	if accent, found, err := profile.Accent(); err != nil || !found ||
		accent.String() != "accent_insensitive_v1" {
		t.Fatal("accent companion differs")
	}
	for _, test := range []struct {
		casefold TextCasefoldProfile
		accent   TextAccentProfile
	}{
		{TextCasefoldDisabled, TextAccentInsensitiveV1},
		{TextCasefoldProfile(99), TextAccentDisabled},
		{TextCasefoldUnicodeDefaultV1, TextAccentProfile(99)},
	} {
		if _, err := NewTextProfile(test.casefold, test.accent); err == nil {
			t.Fatalf("invalid companion combination accepted: %#v", test)
		}
	}
}

func TestTextProfileFailsClosedAndRedacts(t *testing.T) {
	profile, _ := NewTextProfile(TextCasefoldUnicodeDefaultV1, TextAccentDisabled)
	if fmt.Sprint(profile) != queryValueRedaction || fmt.Sprintf("%#v", profile) != queryValueRedaction {
		t.Fatal("text profile formatting leaked")
	}
	for _, value := range []any{profile, TextCasefoldUnicodeDefaultV1, TextAccentDisabled} {
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		}
	}
	if err := (TextProfile{}).Validate(); err == nil {
		t.Fatal("zero text profile validated")
	}
	zero := TextProfile{}
	checks := []func() error{
		func() error { _, err := zero.ExactCollation(); return err },
		func() error { _, err := zero.ExactNormalization(); return err },
		func() error { _, _, err := zero.Casefold(); return err },
		func() error { _, _, err := zero.Accent(); return err },
	}
	for _, check := range checks {
		if check() == nil {
			t.Fatal("zero text accessor succeeded")
		}
	}
	profile.casefold = TextCasefoldProfile(99)
	if err := profile.Validate(); err == nil {
		t.Fatal("tampered text profile validated")
	}
	if _, err := profile.ExactNormalization(); err == nil {
		t.Fatal("tampered text accessor succeeded")
	}
	if _, _, err := profile.Accent(); err == nil {
		t.Fatal("tampered companion accessor succeeded")
	}
}
