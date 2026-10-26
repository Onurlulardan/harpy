package smtp

import "testing"

func TestLifecycleMessageTenantSlugValidation(t *testing.T) {
	for _, testCase := range []struct {
		slug  string
		valid bool
	}{
		{slug: "harpy", valid: true},
		{slug: "tenant_a", valid: true},
		{slug: "", valid: false},
		{slug: " tenant", valid: false},
		{slug: "tenant\nother", valid: false},
	} {
		if validTenantSlug(testCase.slug) != testCase.valid {
			t.Fatalf("tenant slug %q validity mismatch", testCase.slug)
		}
	}
}
