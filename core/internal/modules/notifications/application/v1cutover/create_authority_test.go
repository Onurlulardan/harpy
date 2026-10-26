package v1cutover

import "testing"

func TestCreateAuthorityRejectsUnknownOrigin(t *testing.T) {
	application, _ := NewApplication(cutoverBinderFake{&cutoverStoreFake{}})
	if _, err := application.CreateAuthority(t.Context(), cutoverTestScope(), Origin("other")); err != ErrInvalid {
		t.Fatalf("CreateAuthority() error = %v", err)
	}
}
