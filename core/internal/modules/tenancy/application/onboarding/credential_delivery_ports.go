package onboarding

import (
	"context"
	"regexp"
	"time"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	"github.com/google/uuid"
)

type RecoveryRequest struct {
	OperationID, CurrentAttemptID, CurrentCredentialID uuid.UUID
	NextAttemptID, NextCredentialID                    uuid.UUID
	CurrentCredentialVersion, NextCredentialVersion    int64
	ExpectedRevision                                   int64
	RecoveryTokenSHA256, DeliveryManifestSHA256        string
	SchemaName                                         string
	LeaseSeconds                                       int
}

func (request RecoveryRequest) valid() bool {
	return request.OperationID.Version() == uuid.Version(7) &&
		request.CurrentAttemptID != uuid.Nil && request.NextAttemptID != uuid.Nil &&
		request.CurrentAttemptID != request.NextAttemptID &&
		request.CurrentCredentialID != uuid.Nil && request.NextCredentialID != uuid.Nil &&
		request.CurrentCredentialID != request.NextCredentialID &&
		request.CurrentCredentialVersion > 0 &&
		request.NextCredentialVersion == request.CurrentCredentialVersion+1 &&
		request.ExpectedRevision > 0 && digestPattern.MatchString(request.RecoveryTokenSHA256) &&
		request.DeliveryManifestSHA256 == credentialdeliveryv1.ManifestSHA256() &&
		recoverySchemaPattern.MatchString(request.SchemaName) &&
		request.LeaseSeconds >= 10 && request.LeaseSeconds <= 300
}

type RecoveryBinding struct {
	OperationID, TenantID, FirstOwnerUserID         uuid.UUID
	CurrentCredentialID, NextCredentialID           uuid.UUID
	CurrentAttemptID, NextAttemptID                 uuid.UUID
	CurrentCredentialVersion, NextCredentialVersion int64
	Cause                                           credentialdeliveryv1.Cause
	ManifestSHA256                                  string
}

type RecoveryClaim struct {
	Binding     RecoveryBinding
	Revision    int64
	DatabaseNow time.Time
	LeaseUntil  time.Time
}

type RecoveryRecord struct {
	Binding             RecoveryBinding
	ExpectedRevision    int64
	RecoveryTokenSHA256 string
}

type RecoveryTransition struct {
	State       string
	Revision    int64
	DatabaseNow time.Time
}

type RecoveryRow interface{ Scan(...any) error }

type RecoveryOwnerExecutor struct {
	QueryRow func(context.Context, string, ...any) RecoveryRow
}

type RecoveryTransaction interface {
	Acquire(context.Context, RecoveryRequest) (RecoveryClaim, error)
	BindTenantSchema(context.Context, string) error
	BindIAM() (RecoveryOwnerExecutor, error)
	PersistRecovery(context.Context, RecoveryRecord) (RecoveryTransition, error)
}

type RecoveryUnitOfWork interface {
	Run(context.Context, func(context.Context, RecoveryTransaction) error) error
}

type RecoveryCallback func(
	context.Context,
	RecoveryBinding,
	RecoveryOwnerExecutor,
) error

var recoverySchemaPattern = regexp.MustCompile(`^t_[a-z2-7]{26}$`)
