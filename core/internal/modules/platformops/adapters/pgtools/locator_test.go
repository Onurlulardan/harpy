package pgtools

import "testing"

func TestLocatorSHA256ExcludesPasswordAndCanonicalizesQuery(t *testing.T) {
	left, err := LocatorSHA256(
		"postgresql://harpy_restore:first@db:5432/live?sslmode=require&connect_timeout=5",
	)
	if err != nil {
		t.Fatal(err)
	}
	right, err := LocatorSHA256(
		"postgresql://harpy_restore:second@db:5432/live?connect_timeout=5&sslmode=require",
	)
	if err != nil {
		t.Fatal(err)
	}
	if left != right || len(left) != 64 {
		t.Fatalf("locator digest mismatch: %q %q", left, right)
	}
	scratch, err := LocatorSHA256(
		"postgresql://harpy_restore:second@db:5432/scratch?connect_timeout=5&sslmode=require",
	)
	if err != nil || scratch == left {
		t.Fatalf("distinct database locator was not bound: %q %v", scratch, err)
	}
}
