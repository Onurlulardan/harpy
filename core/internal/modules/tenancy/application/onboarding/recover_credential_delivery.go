package onboarding

import (
	"context"
	"reflect"
	"time"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	credentialdelivery "github.com/Onurlulardan/harpy/core/internal/capabilities/credentialdelivery/api"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func RecoverCredentialDelivery(
	ctx context.Context,
	unit RecoveryUnitOfWork,
	request RecoveryRequest,
	callback RecoveryCallback,
) error {
	if ctx == nil || ctx.Err() != nil || nilRecoveryValue(unit) ||
		callback == nil || !request.valid() {
		return tenantapi.ErrOnboardingInvalid
	}
	return unit.Run(ctx, func(scoped context.Context, transaction RecoveryTransaction) error {
		if nilRecoveryValue(transaction) {
			return tenantapi.ErrOnboardingUnavailable
		}
		claim, err := transaction.Acquire(scoped, request)
		if err != nil {
			return err
		}
		if !claim.valid(request) {
			return tenantapi.ErrOnboardingUnavailable
		}
		if err = transaction.BindTenantSchema(scoped, request.SchemaName); err != nil {
			return err
		}
		iam, err := transaction.BindIAM()
		if err != nil || iam.QueryRow == nil {
			return tenantapi.ErrOnboardingUnavailable
		}
		return credentialdelivery.Run(claim.grant(), claim.DatabaseNow,
			func(scope credentialdelivery.Scope) error {
				if consumeErr := scope.Consume(
					credentialdeliveryv1.SupersedeTemporaryCredential,
				); consumeErr != nil {
					return consumeErr
				}
				if callbackErr := callback(scoped, claim.Binding, iam); callbackErr != nil {
					return callbackErr
				}
				if consumeErr := scope.Consume(
					credentialdeliveryv1.RecordDeliveryAttempt,
				); consumeErr != nil {
					return consumeErr
				}
				result, persistErr := transaction.PersistRecovery(scoped, RecoveryRecord{
					Binding: claim.Binding, ExpectedRevision: claim.Revision,
					RecoveryTokenSHA256: request.RecoveryTokenSHA256,
				})
				if persistErr != nil {
					return persistErr
				}
				if !result.valid() {
					return tenantapi.ErrOnboardingUnavailable
				}
				return nil
			})
	})
}

func (claim RecoveryClaim) valid(request RecoveryRequest) bool {
	binding := claim.Binding
	return binding.OperationID == request.OperationID &&
		binding.CurrentAttemptID == request.CurrentAttemptID &&
		binding.NextAttemptID == request.NextAttemptID &&
		binding.CurrentCredentialID == request.CurrentCredentialID &&
		binding.NextCredentialID == request.NextCredentialID &&
		binding.CurrentCredentialVersion == request.CurrentCredentialVersion &&
		binding.NextCredentialVersion == request.NextCredentialVersion &&
		binding.ManifestSHA256 == credentialdeliveryv1.ManifestSHA256() &&
		validRecoveryCause(binding.Cause) && claim.Revision > 0 &&
		validRecoveryTime(claim.DatabaseNow) && validRecoveryTime(claim.LeaseUntil) &&
		claim.LeaseUntil.After(claim.DatabaseNow)
}

func (claim RecoveryClaim) grant() credentialdelivery.Grant {
	binding := claim.Binding
	return credentialdelivery.Grant{Binding: credentialdelivery.Binding{
		OperationID: binding.OperationID, TenantID: binding.TenantID,
		FirstOwnerUserID:         binding.FirstOwnerUserID,
		CurrentCredentialID:      binding.CurrentCredentialID,
		NextCredentialID:         binding.NextCredentialID,
		CurrentCredentialVersion: binding.CurrentCredentialVersion,
		NextCredentialVersion:    binding.NextCredentialVersion,
		CurrentAttemptID:         binding.CurrentAttemptID, NextAttemptID: binding.NextAttemptID,
		Cause: binding.Cause, ManifestSHA256: binding.ManifestSHA256,
	}, Revision: claim.Revision, LeaseUntil: claim.LeaseUntil}
}

func (transition RecoveryTransition) valid() bool {
	return transition.State == "ready_to_display" && transition.Revision > 0 &&
		validRecoveryTime(transition.DatabaseNow)
}

func validRecoveryCause(cause credentialdeliveryv1.Cause) bool {
	return cause == credentialdeliveryv1.CauseClaimLeaseExpired ||
		cause == credentialdeliveryv1.CauseOutcomeUnknown
}

func validRecoveryTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC
}

func nilRecoveryValue(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
