package licenseauthority

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestLoadLicenseRuntimeEntitlementsDecodesTypedProjection(t *testing.T) {
	environmentID := uuid.MustParse("01900000-0000-7000-8000-000000000501")
	licenseID := uuid.New()
	storage, err := NewStore(DatabaseAccess{QueryRow: func(
		context.Context, string, ...any,
	) Row {
		return rowFixture(func(values ...any) error {
			*values[0].(*string) = `[{"license_id":"` + licenseID.String() +
				`","environment_id":"` + environmentID.String() +
				`","snapshot_sequence":3,"snapshot_sha256":"` +
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				`","state":"grace","entitlement_key":"api.access",` +
				`"value_type":"boolean","boolean_value":true,"integer_value":null,` +
				`"unlimited":false,"channel":"stable",` +
				`"database_now":"2026-07-22T12:00:00Z"}]`
			return nil
		})
	}})
	if err != nil {
		t.Fatal(err)
	}
	grants, err := storage.LoadLicenseRuntimeEntitlements(t.Context(), environmentID)
	if err != nil || len(grants) != 1 || grants[0].LicenseID != licenseID ||
		grants[0].BooleanValue == nil || !*grants[0].BooleanValue {
		t.Fatalf("runtime grants=%+v err=%v", grants, err)
	}
}
