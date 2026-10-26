package cutoverv1

import "testing"

func TestValidateFinalizerLeasePinsLiveStateAndRawEvidenceABI(t *testing.T) {
	value := validTypedFinalizerLease(
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	)
	if err := ValidateFinalizerLease(value); err != nil {
		t.Fatalf("valid finalizer lease rejected: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*FinalizerLease)
	}{
		{"missing state digest", func(v *FinalizerLease) { v.FinalizationStateSHA256 = "" }},
		{"malformed state digest", func(v *FinalizerLease) { v.FinalizationStateSHA256 = "A" }},
		{"old fifteen argument ABI", func(v *FinalizerLease) {
			v.FinalizationProcedureSignature =
				"cutover_authority.finalize_operation_v1(uuid,uuid,uuid,text,bytea," +
					"text,bytea,text,bytea,text,bytea,text,bytea,text,text)"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := value
			test.mutate(&candidate)
			if err := ValidateFinalizerLease(candidate); err == nil {
				t.Fatal("unsafe finalizer lease accepted")
			}
		})
	}
}
