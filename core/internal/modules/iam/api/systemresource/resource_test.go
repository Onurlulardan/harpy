package systemresource

import "testing"

func TestResourceSetIsClosed(t *testing.T) {
	for _, value := range []string{"users", "roles", "organizations"} {
		resource, valid := ParseResource(value)
		key, err := resource.AuthorityResource()
		if !valid || err != nil || key != "iam."+value {
			t.Fatalf("resource %q = %q valid=%v error=%v", value, key, valid, err)
		}
	}
	for _, value := range []string{"", "user", "forms", "USERS"} {
		if _, valid := ParseResource(value); valid {
			t.Fatalf("unknown resource accepted: %q", value)
		}
	}
}
