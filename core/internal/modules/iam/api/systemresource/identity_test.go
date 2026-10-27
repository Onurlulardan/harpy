package systemresource

import "testing"

func mustSystemResourceID(t testing.TB, value string) ID {
	t.Helper()
	id, err := ParseID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
