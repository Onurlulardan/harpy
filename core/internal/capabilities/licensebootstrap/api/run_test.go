package api

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	licensebootstrapv1 "github.com/Onurlulardan/harpy/contracts/licensebootstrap/v1"
	"github.com/google/uuid"
)

func TestRunAllowsOneExactCommandAndClosesScope(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	grant := validGrant(t, now)
	var captured Scope
	err := Run(grant, now, func(scope Scope) error {
		captured = scope
		if scope.Binding() != grant.Binding || scope.Revision() != grant.Revision {
			return ErrInvalid
		}
		return scope.Consume(licensebootstrapv1.PublishLicenseDerivedOffering)
	})
	_, marshalErr := json.Marshal(captured)
	if err != nil || marshalErr == nil || captured.Binding() != (Binding{}) ||
		!errors.Is(captured.Consume(licensebootstrapv1.PublishLicenseDerivedOffering), ErrClosed) {
		t.Fatalf("closed scope drifted: run=%v marshal=%v", err, marshalErr)
	}
}

func TestRunRejectsWrongReplayMissingAndExpiredCommand(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	grant := validGrant(t, now)
	if err := Run(grant, now, func(scope Scope) error {
		return scope.Consume("iam.create_owner")
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
	return Grant{Revision: 2, LeaseUntil: now.Add(time.Minute), Binding: Binding{
		OperationID: newV7(), EnvironmentID: newV7(), LicenseID: uuid.New(),
		LicenseeOrganizationID: uuid.New(), DeploymentMode: "onpremise",
		DeliveryKind: "connected", Issuer: "forgeguard", Sequence: 4,
		SnapshotSHA256: strings.Repeat("1", 64), OfferingVersionID: newV7(),
		Revision: 1, OfferingSHA256: strings.Repeat("2", 64), EffectiveAt: now,
		ManifestSHA256: licensebootstrapv1.ManifestSHA256(),
	}}
}
