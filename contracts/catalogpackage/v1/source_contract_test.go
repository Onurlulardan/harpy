package catalogpackagev1

import "testing"

func TestFirstPartySourceContractRoundTrip(t *testing.T) {
	for _, moduleSlug := range []string{
		"hr", "sales-crm", "accounting", "inventory", "project",
	} {
		contract, err := FirstPartySourceContract(moduleSlug)
		if err != nil {
			t.Fatalf("%s contract: %v", moduleSlug, err)
		}
		parsed, err := ParseFirstPartySourceContract(contract)
		if err != nil || parsed != moduleSlug ||
			!ValidFirstPartySourceContract(contract) {
			t.Fatalf(
				"%s round trip parsed=%q valid=%t err=%v",
				moduleSlug, parsed, ValidFirstPartySourceContract(contract), err,
			)
		}
	}
}

func TestFirstPartySourceContractRejectsNonCanonicalValues(t *testing.T) {
	for _, value := range []string{
		"", "harpy-v1.enterprise-seed.", "enterprise-seed.hr",
		"harpy-v1.enterprise-seed.HR",
		"harpy-v1.enterprise-seed.-hr",
		"harpy-v1.enterprise-seed.hr-",
		"harpy-v1.enterprise-seed.hr--payroll",
		"harpy-v1.enterprise-seed.hr/payroll",
	} {
		if ValidFirstPartySourceContract(value) {
			t.Fatalf("invalid source contract %q was accepted", value)
		}
	}
}
