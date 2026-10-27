package onboarding

import (
	"context"
	"errors"
	"testing"
	"time"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	"github.com/google/uuid"
)

func TestRecoverCredentialDeliveryConsumesBothOwnersInExactOrder(t *testing.T) {
	request, claim := credentialRecoveryFixture()
	transaction := &credentialRecoveryTransactionFixture{claim: claim}
	unit := &credentialRecoveryUnitFixture{transaction: transaction}
	err := RecoverCredentialDelivery(context.Background(), unit, request, func(
		_ context.Context, binding RecoveryBinding, owner RecoveryOwnerExecutor,
	) error {
		if owner.QueryRow == nil {
			t.Fatal("IAM owner was not bound")
		}
		if binding.OperationID != request.OperationID {
			t.Fatal("recovery binding drifted")
		}
		return nil
	})
	if err != nil || !transaction.recorded || transaction.schema != request.SchemaName {
		t.Fatalf("recover err=%v recorded=%t schema=%s", err,
			transaction.recorded, transaction.schema)
	}
}

func TestRecoverCredentialDeliveryRollsBackBeforeTenancyRecordWhenIAMFails(t *testing.T) {
	request, claim := credentialRecoveryFixture()
	transaction := &credentialRecoveryTransactionFixture{claim: claim}
	want := errors.New("IAM unavailable")
	err := RecoverCredentialDelivery(context.Background(),
		&credentialRecoveryUnitFixture{transaction: transaction}, request,
		func(context.Context, RecoveryBinding, RecoveryOwnerExecutor) error {
			return want
		})
	if !errors.Is(err, want) || transaction.recorded {
		t.Fatalf("recover err=%v recorded=%t", err, transaction.recorded)
	}
}

type credentialRecoveryUnitFixture struct {
	transaction *credentialRecoveryTransactionFixture
}

func (unit *credentialRecoveryUnitFixture) Run(
	ctx context.Context, callback func(context.Context, RecoveryTransaction) error,
) error {
	return callback(ctx, unit.transaction)
}

type credentialRecoveryTransactionFixture struct {
	claim    RecoveryClaim
	schema   string
	recorded bool
}

func (fixture *credentialRecoveryTransactionFixture) Acquire(
	context.Context, RecoveryRequest,
) (RecoveryClaim, error) {
	return fixture.claim, nil
}

func (fixture *credentialRecoveryTransactionFixture) BindTenantSchema(
	_ context.Context, schema string,
) error {
	fixture.schema = schema
	return nil
}

func (*credentialRecoveryTransactionFixture) BindIAM() (RecoveryOwnerExecutor, error) {
	return RecoveryOwnerExecutor{QueryRow: func(
		context.Context, string, ...any,
	) RecoveryRow {
		return nil
	}}, nil
}

func (fixture *credentialRecoveryTransactionFixture) PersistRecovery(
	_ context.Context, record RecoveryRecord,
) (RecoveryTransition, error) {
	fixture.recorded = record.Binding == fixture.claim.Binding
	return RecoveryTransition{State: "ready_to_display",
		Revision: fixture.claim.Revision + 1, DatabaseNow: fixture.claim.DatabaseNow}, nil
}

func credentialRecoveryFixture() (RecoveryRequest, RecoveryClaim) {
	operationID := uuid.MustParse("019bfcc1-8362-7a01-8757-6c7046864c89")
	request := RecoveryRequest{OperationID: operationID,
		CurrentAttemptID:         uuid.MustParse("019bfcc1-8362-7a02-8757-6c7046864c89"),
		NextAttemptID:            uuid.MustParse("019bfcc1-8362-7a03-8757-6c7046864c89"),
		CurrentCredentialID:      uuid.MustParse("019bfcc1-8362-7a04-8757-6c7046864c89"),
		NextCredentialID:         uuid.MustParse("019bfcc1-8362-7a05-8757-6c7046864c89"),
		CurrentCredentialVersion: 1, NextCredentialVersion: 2, ExpectedRevision: 3,
		RecoveryTokenSHA256:    repeatDigest("a"),
		DeliveryManifestSHA256: credentialdeliveryv1.ManifestSHA256(),
		SchemaName:             "t_aaaaaaaaaaaaaaaaaaaaaaaaaa", LeaseSeconds: 60}
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	binding := RecoveryBinding{OperationID: request.OperationID,
		TenantID:                 uuid.MustParse("019bfcc1-8362-7a06-8757-6c7046864c89"),
		FirstOwnerUserID:         uuid.MustParse("019bfcc1-8362-7a07-8757-6c7046864c89"),
		CurrentCredentialID:      request.CurrentCredentialID,
		NextCredentialID:         request.NextCredentialID,
		CurrentCredentialVersion: request.CurrentCredentialVersion,
		NextCredentialVersion:    request.NextCredentialVersion,
		CurrentAttemptID:         request.CurrentAttemptID,
		NextAttemptID:            request.NextAttemptID,
		Cause:                    credentialdeliveryv1.CauseOutcomeUnknown,
		ManifestSHA256:           credentialdeliveryv1.ManifestSHA256()}
	return request, RecoveryClaim{Binding: binding, Revision: 4,
		DatabaseNow: now, LeaseUntil: now.Add(time.Minute)}
}

func repeatDigest(value string) string {
	result := ""
	for len(result) < 64 {
		result += value
	}
	return result
}
