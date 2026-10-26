package discountcatalog

import (
	"strings"
	"testing"
)

func TestBuildPresentationPreservesTypedFields(t *testing.T) {
	presentation, err := BuildPresentation(PolicyInput{
		Name: "Summer save", Description: "For loyal customers",
	})
	if err != nil || presentation.Name() != "Summer save" ||
		presentation.Description() != "For loyal customers" {
		t.Fatalf("presentation=%#v err=%v", presentation, err)
	}
}

func TestBuildPresentationRejectsInvalidFields(t *testing.T) {
	invalidUTF8 := string([]byte{0xff})
	cases := []PolicyInput{
		{Name: ""},
		{Name: " padded "},
		{Name: strings.Repeat("x", 101)},
		{Name: invalidUTF8},
		{Name: "Valid", Description: strings.Repeat("x", 4097)},
		{Name: "Valid", Description: invalidUTF8},
	}
	for _, input := range cases {
		if _, err := BuildPresentation(input); err != ErrRequest {
			t.Fatalf("input=%#v err=%v", input, err)
		}
	}
}
