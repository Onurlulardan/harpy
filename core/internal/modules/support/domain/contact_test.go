package domain

import "testing"

func TestNormalizeContactPreservesOnlyBoundedCanonicalValues(t *testing.T) {
	input, err := NormalizeContact(ContactInput{Name: " Ada Lovelace ",
		Email: " ADA@Example.COM ", Company: " Forge ", Phone: " +90 ",
		EmployeeCount: "11-50", Message: " Yardım ", ContactType: "demo_request",
		Source: " landing "})
	if err != nil {
		t.Fatal(err)
	}
	if input.Name != "Ada Lovelace" || input.Email != "ada@example.com" ||
		input.Message != "Yardım" || input.Source != "landing" {
		t.Fatalf("unexpected normalized contact: %#v", input)
	}
	for _, invalid := range []ContactInput{
		{Name: "Ada", Email: "invalid", Message: "help", ContactType: "general"},
		{Name: "Ada", Email: "ada@example.com", Message: "help", ContactType: "unknown"},
		{Name: "Ada", Email: "ada@example.com", Message: "help", ContactType: "general", EmployeeCount: "1000"},
	} {
		if _, err = NormalizeContact(invalid); err == nil {
			t.Fatalf("invalid contact was accepted: %#v", invalid)
		}
	}
}
