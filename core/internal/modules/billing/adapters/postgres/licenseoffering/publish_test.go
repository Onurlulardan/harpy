package licenseoffering

import (
	"context"
	"testing"

	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseoffering"
	"github.com/google/uuid"
)

type rowFixture func(...any) error

func (fixture rowFixture) Scan(destinations ...any) error {
	return fixture(destinations...)
}

func TestPublishLicenseOfferingScansExactReceipt(t *testing.T) {
	operationID := uuid.MustParse("01900000-0000-7000-8000-000000000201")
	versionID := uuid.MustParse("01900000-0000-7000-8000-000000000202")
	digest := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	storage, err := NewStore(DatabaseAccess{QueryRow: func(
		context.Context, string, ...any,
	) Row {
		return rowFixture(func(values ...any) error {
			*values[0].(*uuid.UUID), *values[1].(*uuid.UUID) = operationID, versionID
			*values[2].(*int64), *values[3].(*string) = 1, digest
			*values[4].(*bool) = false
			return nil
		})
	}})
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := storage.PublishLicenseOffering(t.Context(), offeringapp.PublishRequest{})
	if err != nil || receipt.OperationID != operationID ||
		receipt.OfferingVersionID != versionID || receipt.OfferingSHA256 != digest {
		t.Fatalf("receipt = %+v, %v", receipt, err)
	}
}
