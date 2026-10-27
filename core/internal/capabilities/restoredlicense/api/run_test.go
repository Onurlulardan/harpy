package api

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	restoredlicensev1 "github.com/Onurlulardan/harpy/contracts/restoredlicense/v1"
	"github.com/google/uuid"
)

func TestRunAllowsOneExactRestoreCommandAndCloses(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	grant := validGrant(t, now)
	var captured Scope
	err := Run(grant, now, func(scope Scope) error {
		captured = scope
		return scope.Consume(restoredlicensev1.ReconcileRestoredLicense)
	})
	_, marshalErr := json.Marshal(captured)
	if err != nil || marshalErr == nil || captured.Binding() != (Binding{}) ||
		!errors.Is(captured.Consume(restoredlicensev1.ReconcileRestoredLicense), ErrClosed) {
		t.Fatalf("closed restore scope drifted: run=%v marshal=%v", err, marshalErr)
	}
}

func TestRunRejectsWrongReplayMissingAndExpiredCommand(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	grant := validGrant(t, now)
	if err := Run(grant, now, func(scope Scope) error {
		return scope.Consume("iam.restore_owner")
	}); !errors.Is(err, ErrDenied) {
		t.Fatalf("wrong command accepted: %v", err)
	}
	if err := Run(grant, now, func(Scope) error { return nil }); !errors.Is(err, ErrOrder) {
		t.Fatalf("missing command accepted: %v", err)
	}
	grant.LeaseUntil = now
	if err := Run(grant, now, func(Scope) error { return nil }); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expired grant accepted: %v", err)
	}
}

func validGrant(t *testing.T, now time.Time) Grant {
	t.Helper()
	newV7 := func() uuid.UUID {
		value, err := uuid.NewV7()
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	return Grant{Revision: 3, LeaseUntil: now.Add(time.Minute), Binding: Binding{
		OperationID: newV7(), EnvironmentID: uuid.New(), RestoreReceiptID: uuid.New(),
		RestoreReceiptSHA256: strings.Repeat("1", 64), LicenseBindingID: uuid.New(),
		LicenseID: uuid.New(), LicenseeOrganizationID: uuid.New(), DeliveryKind: "connected",
		Issuer: "harpy", PreviousSequence: 8, PreviousSHA256: strings.Repeat("2", 64),
		NextSequence: 9, NextSHA256: strings.Repeat("3", 64),
		ManifestSHA256: restoredlicensev1.ManifestSHA256(),
	}}
}
