package formsrecordpagev1

import "testing"

func TestLimitsFreezeADR0100Contract(t *testing.T) {
	if MaximumRows != 1000 || MaximumCanonicalBytes != 8*1024*1024 ||
		MaximumRowBytes != 1024*1024 {
		t.Fatal("ADR-0100 Forms record page contract drifted")
	}
}
