package api

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	"github.com/google/uuid"
)

func TestRunConsumesExactOrderedCredentialRecoveryAndCloses(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	grant := validGrant(t, now)
	var captured Scope
	err := Run(grant, now, func(scope Scope) error {
		captured = scope
		if err := scope.Consume(credentialdeliveryv1.SupersedeTemporaryCredential); err != nil {
			return err
		}
		return scope.Consume(credentialdeliveryv1.RecordDeliveryAttempt)
	})
	_, marshalErr := json.Marshal(captured)
	if err != nil || marshalErr == nil || captured.Binding() != (Binding{}) ||
		!errors.Is(captured.Consume(credentialdeliveryv1.RecordDeliveryAttempt), ErrClosed) {
		t.Fatalf("closed credential scope drifted: run=%v marshal=%v", err, marshalErr)
	}
}

func TestRunRejectsWrongOrderReplayMissingAcknowledgedAndExpiry(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	grant := validGrant(t, now)
	if err := Run(grant, now, func(scope Scope) error {
		return scope.Consume(credentialdeliveryv1.RecordDeliveryAttempt)
	}); !errors.Is(err, ErrOrder) {
		t.Fatalf("wrong order accepted: %v", err)
	}
	if err := Run(grant, now, func(scope Scope) error {
		return scope.Consume("installation.acknowledge_credential")
	}); !errors.Is(err, ErrDenied) {
		t.Fatalf("acknowledged mutation accepted: %v", err)
	}
	grant.LeaseUntil = now
	if err := Run(grant, now, func(Scope) error { return nil }); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expired scope accepted: %v", err)
	}
}

func validGrant(t *testing.T, now time.Time) Grant {
	t.Helper()
	operationID, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return Grant{Revision: 6, LeaseUntil: now.Add(time.Minute), Binding: Binding{
		OperationID: operationID, TenantID: uuid.New(), FirstOwnerUserID: uuid.New(),
		CurrentCredentialID: uuid.New(), CurrentCredentialVersion: 2,
		NextCredentialID: uuid.New(), NextCredentialVersion: 3,
		CurrentAttemptID: uuid.New(), NextAttemptID: uuid.New(),
		Cause:          credentialdeliveryv1.CauseOutcomeUnknown,
		ManifestSHA256: credentialdeliveryv1.ManifestSHA256(),
	}}
}
