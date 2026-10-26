package authorization

import "testing"

func TestNewRequestUsesClosedCatalogKeysAndCurrentVersions(t *testing.T) {
	valid := []string{"iam.users", "forms.contact-form", "work.task_read"}
	for _, key := range valid {
		if _, err := NewRequest(key, "read", 1, 1); err != nil {
			t.Fatalf("valid key %q: %v", key, err)
		}
	}
	invalid := []string{"", "IAM.Users", ".iam", "iam..users", "iam/users", "iam users"}
	for _, key := range invalid {
		if _, err := NewRequest(key, "read", 1, 1); err == nil {
			t.Fatalf("invalid key %q was accepted", key)
		}
	}
	if _, err := NewRequest("iam.users", "read", 0, 1); err == nil {
		t.Fatal("zero token version was accepted")
	}
	if _, err := NewRequest("iam.users", "read", 1, 0); err == nil {
		t.Fatal("zero authorization version was accepted")
	}
}
