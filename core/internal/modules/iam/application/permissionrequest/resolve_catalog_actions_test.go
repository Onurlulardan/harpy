package permissionrequest

import "testing"

func TestResolveCatalogActionsUsesCanonicalIAMKeyGrammar(t *testing.T) {
	for _, value := range []string{"read", "read_own", "documents.read-all"} {
		if !validCatalogKey(value) {
			t.Fatalf("valid action key rejected: %s", value)
		}
	}
	for _, value := range []string{"", "Read", ".read", "read..own", "read own", "read-"} {
		if validCatalogKey(value) {
			t.Fatalf("invalid action key accepted: %q", value)
		}
	}
}
