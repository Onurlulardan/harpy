package transaction

import "testing"

func TestCredentialLookupConstructorsSealExactKeyShape(t *testing.T) {
	login, err := NewLoginCredentialLookup(
		"iam.login", testCredentialTenantID,
		CredentialEmailIdentifier, testCredentialKeyID, [32]byte{1, 2},
	)
	if err != nil || login.Purpose() != CredentialLogin ||
		login.LocatorKind() != CredentialLoginDirectory || len(login.Key()) != 32 ||
		login.IdentifierKind() != CredentialEmailIdentifier ||
		login.KeyID() != testCredentialKeyID || login.TenantID() != testCredentialTenantID {
		t.Fatalf("login lookup = %+v, %v", login, err)
	}
	initial, err := NewInitialPasswordCredentialLookup(
		"iam.initial-password.start", testCredentialTenantID,
		CredentialEmailIdentifier, testCredentialKeyID, [32]byte{8},
	)
	if err != nil || initial.Purpose() != CredentialInitialPassword ||
		initial.LocatorKind() != CredentialLoginDirectory {
		t.Fatalf("initial password lookup = %+v, %v", initial, err)
	}
	refresh, err := NewRefreshCredentialLookup("iam.refresh", [16]byte{3, 4})
	if err != nil || refresh.Purpose() != CredentialRefresh ||
		refresh.LocatorKind() != CredentialRefreshLocator || len(refresh.Key()) != 16 {
		t.Fatalf("refresh lookup = %+v, %v", refresh, err)
	}
	route, err := NewTenantRouteCredentialLookup(
		"iam.invitation", CredentialInvitation, testCredentialTenantID, [32]byte{5},
	)
	if err != nil || route.Purpose() != CredentialInvitation ||
		route.LocatorKind() != CredentialTenantRoute || len(route.Key()) != 32 {
		t.Fatalf("route lookup = %+v, %v", route, err)
	}
	session, err := NewSessionCredentialLookup(
		"iam.reauth", testCredentialTenantID, [32]byte{6},
	)
	if err != nil || session.Purpose() != CredentialReauthentication ||
		session.LocatorKind() != CredentialSessionLocator || len(session.Key()) != 32 {
		t.Fatalf("session lookup = %+v, %v", session, err)
	}
	key := login.Key()
	key[0] = 99
	if login.Key()[0] != 1 {
		t.Fatal("lookup key escaped by reference")
	}
}

func TestCredentialLookupConstructorsRejectInvalidPurposeAndKeys(t *testing.T) {
	if _, err := NewLoginCredentialLookup(
		"invalid operation", testCredentialTenantID,
		CredentialEmailIdentifier, "key", [32]byte{1},
	); err == nil {
		t.Fatal("invalid operation was accepted")
	}
	if _, err := NewLoginCredentialLookup(
		"iam.login", testCredentialTenantID,
		CredentialEmailIdentifier, "key", [32]byte{},
	); err == nil {
		t.Fatal("zero login key was accepted")
	}
	if _, err := NewLoginCredentialLookup(
		"iam.login", testCredentialTenantID, "phone_v1", "bad key", [32]byte{1},
	); err == nil {
		t.Fatal("invalid login metadata was accepted")
	}
	if _, err := NewLoginCredentialLookup(
		"iam.login", [16]byte{}, CredentialEmailIdentifier, "key", [32]byte{1},
	); err == nil {
		t.Fatal("nil login tenant was accepted")
	}
	if _, err := NewRefreshCredentialLookup("iam.refresh", [16]byte{}); err == nil {
		t.Fatal("zero selector was accepted")
	}
	for _, purpose := range []CredentialPurpose{
		CredentialLogin, CredentialRefresh, CredentialReauthentication, "unknown",
	} {
		if _, err := NewTenantRouteCredentialLookup(
			"iam.route", purpose, testCredentialTenantID, [32]byte{1},
		); err == nil {
			t.Fatalf("tenant route purpose %q was accepted", purpose)
		}
	}
	if (CredentialLookup{}).Key() != nil {
		t.Fatal("zero lookup exposed a key")
	}
}

var testCredentialTenantID = [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

const testCredentialKeyID = "-_v7-_v7-_v7-_v7"
