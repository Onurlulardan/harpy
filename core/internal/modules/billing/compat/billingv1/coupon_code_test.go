package billingv1

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestSecureCouponCodeHasExactIndependentWireShape(t *testing.T) {
	pattern := regexp.MustCompile(`^[A-Z0-9]{4}-[A-Z0-9]{4}$`)
	for range 64 {
		code, err := secureCouponCode()
		if err != nil || !pattern.MatchString(code) {
			t.Fatalf("code=%q error=%v", code, err)
		}
	}
}

func TestCreateCouponRetriesGeneratedCodeConflict(t *testing.T) {
	catalog := &couponRetryCatalog{}
	codes := []string{"AAAA-AAAA", "BBBB-BBBB"}
	handler := &Handler{
		catalog: catalog,
		clock: func() time.Time {
			return time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
		},
		couponCodeSource: func() (string, error) {
			code := codes[0]
			codes = codes[1:]
			return code, nil
		},
	}
	name, kind, amount := "Save", "percentage", int64(10)
	view, err := handler.createCouponView(
		context.Background(), billingapi.Access{}, couponReferenceAccess{}, "key",
		couponRequest{Name: &name, DiscountType: &kind, DiscountValue: &amount}, true,
	)
	if err != nil || view.Code != "BBBB-BBBB" ||
		len(catalog.codes) != 2 || catalog.codes[0] != "AAAA-AAAA" ||
		catalog.codes[1] != "BBBB-BBBB" {
		t.Fatalf("view=%#v codes=%v error=%v", view, catalog.codes, err)
	}
}

type couponRetryCatalog struct {
	billingapi.PlatformCatalog
	codes []string
}

func (catalog *couponRetryCatalog) CreateDiscount(
	_ context.Context,
	command billingapi.CreateDiscountCommand,
) (billingapi.DiscountView, error) {
	catalog.codes = append(catalog.codes, command.InitialVersion.Code)
	if len(catalog.codes) == 1 {
		return billingapi.DiscountView{}, billingapi.ErrConflict
	}
	if command.InitialVersion.Code == "" {
		return billingapi.DiscountView{}, errors.New("empty generated code")
	}
	return billingapi.DiscountView{Code: command.InitialVersion.Code}, nil
}
