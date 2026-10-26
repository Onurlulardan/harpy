package v1cutover

import "testing"

func TestPauseV1RequiresWritableV1Authority(t *testing.T) {
	application, _ := NewApplication(cutoverBinderFake{&cutoverStoreFake{}})
	authority := Authority{Writer: "v2", State: "writable", Generation: 1, RowVersion: 1}
	if _, err := application.PauseV1(t.Context(), cutoverTestScope(), authority); err != ErrInvalid {
		t.Fatalf("PauseV1() error = %v", err)
	}
}
