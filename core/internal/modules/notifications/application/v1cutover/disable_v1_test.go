package v1cutover

import "testing"

func TestDisableV1RejectsWritableAuthority(t *testing.T) {
	application, _ := NewApplication(cutoverBinderFake{&cutoverStoreFake{}})
	authority := Authority{Writer: "v1", State: "writable", Generation: 1, RowVersion: 1}
	if _, err := application.DisableV1(t.Context(), cutoverTestScope(), authority, Evidence{}); err != ErrInvalid {
		t.Fatalf("DisableV1() error = %v", err)
	}
}
