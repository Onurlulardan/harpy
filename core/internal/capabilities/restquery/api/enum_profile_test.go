package api

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

const testEnumVersion = "11111111-1111-4111-8111-111111111111"

func TestEnumProfileRequiresCanonicalNonzeroVersionAndValues(t *testing.T) {
	value := mustEnumValue(t, "OPEN", nil)
	for _, version := range []string{
		"", uuid.Nil.String(), "11111111111141118111111111111111",
		"11111111-1111-4111-8111-11111111111A",
		"11111111-1111-4111-0111-111111111111",
	} {
		if _, err := NewEnumProfile(version, []EnumValue{value}); err == nil {
			t.Fatalf("invalid enum version %q accepted", version)
		}
	}
	if _, err := NewEnumProfile(testEnumVersion, nil); err == nil {
		t.Fatal("empty enum values accepted")
	}
}

func TestEnumProfileEnforcesUniqueCodesAndAllOrNoOrdering(t *testing.T) {
	first, second := 10, 20
	open := mustEnumValue(t, "OPEN", &first)
	closed := mustEnumValue(t, "CLOSED", &second)
	profile, err := NewEnumProfile(testEnumVersion, []EnumValue{open, closed})
	if err != nil {
		t.Fatal(err)
	}
	version, versionErr := profile.Version()
	if versionErr != nil || version.String() != testEnumVersion {
		t.Fatal("enum version changed")
	}
	ordered, orderedErr := profile.Ordered()
	if orderedErr != nil || !ordered {
		t.Fatal("complete ordinals lost ordering")
	}
	if found, ok, err := profile.Lookup("OPEN"); err != nil || !ok || enumCode(found) != "OPEN" {
		t.Fatal("code lookup failed")
	}
	if found, ok, err := profile.LookupOrdinal(20); err != nil ||
		!ok || enumCode(found) != "CLOSED" {
		t.Fatal("ordinal lookup failed")
	}
	duplicateOrdinal := mustEnumValue(t, "OTHER", &first)
	for _, values := range [][]EnumValue{
		{open, open},
		{open, mustEnumValue(t, "NO_ORDINAL", nil)},
		{open, duplicateOrdinal},
	} {
		if _, err := NewEnumProfile(testEnumVersion, values); err == nil {
			t.Fatalf("invalid enum values accepted: %#v", values)
		}
	}
}

func TestEnumValueRequiresStableNFCCode(t *testing.T) {
	for _, code := range []string{"", "bad\x00code", "e\u0301", string([]byte{0xff})} {
		if _, err := NewEnumValue(code, nil); err == nil {
			t.Fatalf("invalid enum code %q accepted", code)
		}
	}
}

func TestEnumProfileOwnsStorageAndRedacts(t *testing.T) {
	values := []EnumValue{mustEnumValue(t, "SECRET", nil)}
	profile, err := NewEnumProfile(testEnumVersion, values)
	if err != nil {
		t.Fatal(err)
	}
	values[0] = EnumValue{}
	returned, valuesErr := profile.Values()
	if valuesErr != nil {
		t.Fatal(valuesErr)
	}
	returned[0] = EnumValue{}
	if stored, found, err := profile.Lookup("SECRET"); err != nil ||
		!found || enumCode(stored) != "SECRET" {
		t.Fatal("enum profile exposed mutable storage")
	}
	if fmt.Sprint(profile) != queryValueRedaction || fmt.Sprintf("%#v", profile) != queryValueRedaction {
		t.Fatal("enum profile formatting leaked")
	}
	storedValues, err := profile.Values()
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []any{profile, storedValues[0]} {
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		}
	}
	if err := (EnumProfile{}).Validate(); err == nil {
		t.Fatal("zero enum profile validated")
	}
	assertEnumAccessorsFail(t, EnumProfile{})
	if _, err := (EnumValue{}).Code(); err == nil {
		t.Fatal("zero enum value accessor succeeded")
	}
	if _, _, err := (EnumValue{}).Ordinal(); err == nil {
		t.Fatal("zero enum ordinal accessor succeeded")
	}
	profile.values[0].code = ""
	if err := profile.Validate(); err == nil {
		t.Fatal("tampered enum profile validated")
	}
	assertEnumAccessorsFail(t, profile)
}

func mustEnumValue(t *testing.T, code string, ordinal *int) EnumValue {
	t.Helper()
	value, err := NewEnumValue(code, ordinal)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func enumCode(value EnumValue) string {
	code, _ := value.Code()
	return code
}

func assertEnumAccessorsFail(t *testing.T, profile EnumProfile) {
	t.Helper()
	if _, err := profile.Version(); err == nil {
		t.Fatal("invalid enum version accessor succeeded")
	}
	if _, err := profile.Ordered(); err == nil {
		t.Fatal("invalid enum ordering accessor succeeded")
	}
	if _, err := profile.Values(); err == nil {
		t.Fatal("invalid enum values accessor succeeded")
	}
	if _, _, err := profile.Lookup("CODE"); err == nil {
		t.Fatal("invalid enum lookup succeeded")
	}
	if _, _, err := profile.LookupOrdinal(1); err == nil {
		t.Fatal("invalid enum ordinal lookup succeeded")
	}
}
