package restoredlicense

import (
	"context"
	"strings"
	"testing"
	"time"

	restoredlicensev1 "github.com/Onurlulardan/harpy/contracts/restoredlicense/v1"
	"github.com/google/uuid"
)

type restoredLicenseStoreStub struct {
	result Result
}

func (store restoredLicenseStoreStub) ReconcileRestoredLicense(
	context.Context,
	Input,
) (Result, error) {
	return store.result, nil
}

func TestReconcileAndRunRestoredLicenseGrantStayExactlyBound(t *testing.T) {
	operation := uuid.Must(uuid.NewV7())
	input := Input{
		OperationID: operation, EnvironmentID: uuid.New(),
		RestoreReceiptID:     uuid.New(),
		RestoreReceiptSHA256: strings.Repeat("a", 64),
		LicenseID:            uuid.New(), LicenseeOrganizationID: uuid.New(),
		DeliveryKind: "connected", Issuer: "harpy",
		PreviousSequence: 4, PreviousSHA256: strings.Repeat("b", 64),
		NextSequence: 5, NextSHA256: strings.Repeat("c", 64),
		AuthorityRevision: 7,
	}
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	result := Result{
		Grant: Grant{
			Binding: Binding{
				OperationID:            input.OperationID,
				EnvironmentID:          input.EnvironmentID,
				RestoreReceiptID:       input.RestoreReceiptID,
				RestoreReceiptSHA256:   input.RestoreReceiptSHA256,
				LicenseBindingID:       uuid.Must(uuid.NewV7()),
				LicenseID:              input.LicenseID,
				LicenseeOrganizationID: input.LicenseeOrganizationID,
				DeliveryKind:           input.DeliveryKind, Issuer: input.Issuer,
				PreviousSequence: input.PreviousSequence,
				PreviousSHA256:   input.PreviousSHA256,
				NextSequence:     input.NextSequence, NextSHA256: input.NextSHA256,
				ManifestSHA256: restoredLicenseManifestSHA256(),
			},
			Revision: input.AuthorityRevision, LeaseUntil: now.Add(time.Minute),
		},
		DatabaseNow: now,
	}
	grant, err := Reconcile(
		context.Background(), restoredLicenseStoreStub{result: result}, input,
	)
	if err != nil {
		t.Fatal(err)
	}
	called := false
	if err = Run(grant.Grant, grant.DatabaseNow, func(scope Scope) error {
		called = true
		return scope.Consume(restoredlicensev1.ReconcileRestoredLicense)
	}); err != nil || !called {
		t.Fatalf("run restored license grant called=%v err=%v", called, err)
	}
}

func restoredLicenseManifestSHA256() string {
	return "982f8d8b6315f237e87a80904ceda1ed4ef54d32931907cda9be1db07a2f5d45"
}
