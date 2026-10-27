package cataloghttp

import (
	"net/http/httptest"
	"testing"
)

const firstPartyPackageID = "e3000000-0000-0000-0000-000000000001"

func TestPackageIdentityAcceptsOpaqueCanonicalNonZeroUUID(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/catalog/packages/"+firstPartyPackageID, nil)
	request.SetPathValue("package_id", firstPartyPackageID)
	value, err := PathPackageID(request, "package_id")
	if err != nil || value.String() != firstPartyPackageID {
		t.Fatalf("package path = %s, %v", value, err)
	}
	request = httptest.NewRequest("GET",
		"/api/catalog/packages?limit=1&after="+firstPartyPackageID, nil)
	limit, after, err := PackagePageInput(request)
	if err != nil || limit != 1 || after == nil || after.String() != firstPartyPackageID {
		t.Fatalf("package page = %d/%v, %v", limit, after, err)
	}
}

func TestPackageIdentityRejectsMalformedNilAndNonCanonicalValues(t *testing.T) {
	for _, value := range []string{
		"", "not-a-uuid", "00000000-0000-0000-0000-000000000000",
		"E3000000-0000-0000-0000-000000000001",
		"e3000000000000000000000000000001",
	} {
		t.Run(value, func(t *testing.T) {
			if _, err := ParsePackageID(value); err == nil {
				t.Fatalf("package identity %q was accepted", value)
			}
		})
	}
}

func TestInstallationCursorRemainsUUIDv7Only(t *testing.T) {
	request := httptest.NewRequest("GET",
		"/api/catalog/installations?after="+firstPartyPackageID, nil)
	if _, _, err := PageInput(request); err == nil {
		t.Fatal("opaque package identity was accepted as installation cursor")
	}
}
