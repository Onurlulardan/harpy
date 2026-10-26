package entitlement

import "testing"

func TestKeyVocabulary(t *testing.T) {
	valid := []string{"a", "forms.submissions", "documents_storage", "api.v2_calls"}
	for _, input := range valid {
		key, err := NewKey(input)
		if err != nil || !key.Valid() {
			t.Fatalf("valid key %q rejected: %v", input, err)
		}
	}
	invalid := []string{"", ".forms", "forms.", "Forms", "forms-submissions", " forms"}
	for _, input := range invalid {
		if _, err := NewKey(input); err == nil {
			t.Fatalf("invalid key %q accepted", input)
		}
	}
	tooLong := "a"
	for len(tooLong) < 129 {
		tooLong += "a"
	}
	if _, err := NewKey(tooLong); err == nil {
		t.Fatal("129-byte key accepted")
	}
}
