package billingv1adoptionapp

import "testing"

func TestCompileRejectsMissingAdoptionIdentity(t *testing.T) {
	_, err := Compile(AdoptionSnapshot{}, AdoptionMapping{})
	if err == nil {
		t.Fatal("Compile accepted a snapshot without environment identity")
	}
}
