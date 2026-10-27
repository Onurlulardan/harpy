package discount

import "testing"

func TestNormalizeCodeUsesStrictASCIIUppercase(t *testing.T) {
	got, err := NormalizeCode("launch_2026-x")
	if err != nil || got != "LAUNCH_2026-X" || !got.Valid() {
		t.Fatalf("got %q, %v", got, err)
	}
	for _, value := range []string{"", " SAVE", "SAVE ", "indirim-ı", "é", "A/B"} {
		if _, err := NormalizeCode(value); err != ErrCode {
			t.Fatalf("%q: expected ErrCode, got %v", value, err)
		}
	}
}
