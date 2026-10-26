package billing

import "testing"

func TestCanonicalCatalogIntervalsAreExactUTCMicrosecondsAndHalfOpen(t *testing.T) {
	validEnd := "2026-09-01T00:00:00.000000Z"
	tests := []struct {
		name  string
		start string
		end   *string
		valid bool
	}{
		{name: "open", start: "2026-08-01T00:00:00.000000Z", valid: true},
		{name: "bounded", start: "2026-08-01T00:00:00.000000Z", end: &validEnd, valid: true},
		{name: "no fraction", start: "2026-08-01T00:00:00Z"},
		{name: "offset", start: "2026-08-01T03:00:00.000000+03:00"},
		{name: "empty", start: ""},
	}
	for _, test := range tests {
		if got := validUTCInterval(test.start, test.end); got != test.valid {
			t.Fatalf("%s valid=%t", test.name, got)
		}
	}
	equal := "2026-08-01T00:00:00.000000Z"
	if validUTCInterval(equal, &equal) {
		t.Fatal("empty half-open interval was accepted")
	}
}

func TestDiscountUseLimitsMustBePositiveWhenPresent(t *testing.T) {
	zero, one := int64(0), int64(1)
	if positiveOptional(&zero) || !positiveOptional(&one) || !positiveOptional(nil) {
		t.Fatal("optional positive limit validation mismatch")
	}
}

func TestDiscountOfferingEligibilityIsExplicitUUIDv7XOR(t *testing.T) {
	v7 := "018f47a0-7b9c-7a35-8f31-5f8c0f745101"
	if !validOfferingEligibility(true, nil) ||
		!validOfferingEligibility(false, []string{v7}) {
		t.Fatal("valid explicit applicability was rejected")
	}
	if validOfferingEligibility(false, nil) ||
		validOfferingEligibility(true, []string{v7}) ||
		validOfferingEligibility(false, []string{v7, v7}) ||
		validOfferingEligibility(false, []string{"00000000-0000-4000-8000-000000000001"}) {
		t.Fatal("ambiguous, duplicate, or non-UUIDv7 applicability was accepted")
	}
}

func TestCampaignDiscountVersionsAllowEmptyOrUniqueUUIDv7(t *testing.T) {
	v7 := "018f47a0-7b9c-7a35-8f31-5f8c0f745101"
	if !validUUIDv7Set(nil, 0, 256) || !validUUIDv7Set([]string{v7}, 0, 256) {
		t.Fatal("valid campaign discount set was rejected")
	}
	if validUUIDv7Set([]string{v7, v7}, 0, 256) ||
		validUUIDv7Set([]string{"00000000-0000-4000-8000-000000000001"}, 0, 256) {
		t.Fatal("duplicate or non-UUIDv7 campaign discount set was accepted")
	}
}
