package billingv1adoptionapp

import "testing"

func TestDeterministicIdentityIsStableUUIDV7(t *testing.T) {
	first, err := deterministicIdentity("offering", "source-1")
	if err != nil {
		t.Fatal(err)
	}
	second, err := deterministicIdentity("offering", "source-1")
	if err != nil {
		t.Fatal(err)
	}
	other, err := deterministicIdentity("offering-version", "source-1")
	if err != nil {
		t.Fatal(err)
	}
	if first != second || first == other || first.Version() != 7 {
		t.Fatalf("deterministic identity mismatch: %s %s %s", first, second, other)
	}
}

func TestDeterministicIdentityRejectsEmptyComponents(t *testing.T) {
	if _, err := deterministicIdentity("offering", ""); err == nil {
		t.Fatal("empty identity component was accepted")
	}
}
