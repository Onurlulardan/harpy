package platformlicense

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestReadLicenseDecodesEntitlements(t *testing.T) {
	environmentID, licenseID := testV7(t), testV7(t)
	now := time.Date(2026, 7, 23, 15, 0, 0, 0, time.FixedZone("fixture", 3*3600))
	query := &queryFixture{scan: func(destinations ...any) error {
		*destinations[0].(*uuid.UUID) = licenseID
		*destinations[1].(*uuid.UUID) = environmentID
		*destinations[2].(*int64) = 7
		*destinations[3].(*string) = strings.Repeat("a", 64)
		*destinations[4].(*string) = "valid"
		*destinations[5].(*string) = "stable"
		*destinations[6].(*string) = `{"workflow":true}`
		*destinations[7].(*string) = `{"users":{"unlimited":false,"value":100}}`
		*destinations[8].(*time.Time) = now
		return nil
	}}
	value, err := fixtureStore(t, query).ReadLicense(t.Context(), environmentID)
	if err != nil || value.LicenseID != licenseID || !value.Features["workflow"] ||
		value.Limits["users"].Value != 100 || value.DatabaseNow.Location() != time.UTC {
		t.Fatalf("license=%+v err=%v", value, err)
	}
	if query.statement != readLicenseSQL || query.arguments[0] != environmentID {
		t.Fatalf("statement or arguments mismatch")
	}
}
