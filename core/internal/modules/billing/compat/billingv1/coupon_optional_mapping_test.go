package billingv1

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestCouponNullableFieldsDistinguishAbsentNullAndValueOnUpdate(t *testing.T) {
	oldEnd := "2026-08-01T00:00:00.000000Z"
	oldMax, oldMonths := int64(8), 3
	current := billingapi.DiscountView{
		ID: "discount", Code: "SAVE", Name: "Save", Kind: "basis_points",
		BasisPoints: int64Pointer(1000), Duration: "repeating",
		DurationCycles: &oldMonths, GlobalUseLimit: &oldMax, Available: true,
		AllOfferings: true, ValidStart: "2026-07-01T00:00:00.000000Z",
		ValidEnd: &oldEnd,
	}
	tests := []struct {
		name       string
		document   string
		wantEnd    *string
		wantMax    *int64
		wantMonths *int
	}{
		{name: "absent", document: `{}`, wantEnd: &oldEnd,
			wantMax: &oldMax, wantMonths: &oldMonths},
		{name: "explicit null", document: `{"valid_until":null,"max_uses":null,"duration_months":null}`},
		{name: "replacement", document: `{"valid_until":"2026-09-02T03:04:05Z","max_uses":12,"duration_months":6}`,
			wantEnd: stringPointer("2026-09-02T03:04:05.000000Z"),
			wantMax: int64Pointer(12), wantMonths: intPointer(6)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := decodeCouponMappingRequest(t, test.document)
			version, err := (&Handler{clock: time.Now}).couponVersion(
				context.Background(), current, input, current.Code, couponReferenceAccess{},
			)
			if err != nil {
				t.Fatal(err)
			}
			assertOptionalString(t, "valid_until", version.ValidEnd, test.wantEnd)
			assertOptionalInt64(t, "max_uses", version.GlobalUseLimit, test.wantMax)
			assertOptionalInt(t, "duration_months", version.DurationCycles, test.wantMonths)
		})
	}
}

func TestCouponNullableFieldsMapCreateNullAndAbsentToNil(t *testing.T) {
	for _, document := range []string{
		`{}`,
		`{"valid_until":null,"max_uses":null,"duration_months":null}`,
	} {
		input := decodeCouponMappingRequest(t, document)
		version, err := (&Handler{clock: func() time.Time {
			return time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
		}}).couponVersion(
			context.Background(), billingapi.DiscountView{}, input, "SAVE",
			couponReferenceAccess{},
		)
		if err != nil {
			t.Fatal(err)
		}
		if version.ValidEnd != nil || version.GlobalUseLimit != nil ||
			version.DurationCycles != nil {
			t.Fatalf("document=%s version=%#v", document, version)
		}
	}
}

func decodeCouponMappingRequest(t *testing.T, document string) couponRequest {
	t.Helper()
	var input couponRequest
	if err := json.Unmarshal([]byte(document), &input); err != nil {
		t.Fatal(err)
	}
	return input
}

func assertOptionalString(t *testing.T, field string, got, want *string) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Fatalf("%s=%v want=%v", field, got, want)
		}
		return
	}
	if *got != *want {
		t.Fatalf("%s=%q want=%q", field, *got, *want)
	}
}

func assertOptionalInt64(t *testing.T, field string, got, want *int64) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Fatalf("%s=%v want=%v", field, got, want)
		}
		return
	}
	if *got != *want {
		t.Fatalf("%s=%d want=%d", field, *got, *want)
	}
}

func assertOptionalInt(t *testing.T, field string, got, want *int) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Fatalf("%s=%v want=%v", field, got, want)
		}
		return
	}
	if *got != *want {
		t.Fatalf("%s=%d want=%d", field, *got, *want)
	}
}

func intPointer(value int) *int          { return &value }
func int64Pointer(value int64) *int64    { return &value }
func stringPointer(value string) *string { return &value }
