package dashboardexport

import (
	"context"
	"regexp"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

var failureCodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,63}$`)

func (storage *WorkerStore) MarkOutcomeUnknown(
	ctx context.Context, write exportapp.RenderUnknown,
) error {
	if !validAttemptIdentity(write.Attempt) || write.EffectID.Version() != 7 ||
		write.EffectRequestSHA256 == [32]byte{} || write.ObservedAt.IsZero() {
		return exportapp.ErrInvalid
	}
	var persisted bool
	err := storage.inTenant(ctx, write.Attempt.TenantID,
		func(boundContext context.Context, bound *workerBound) error {
			schema, err := workerSchema(boundContext, bound)
			if err != nil {
				return err
			}
			return bound.queryRow(boundContext, workerUnknownSQL, schema,
				write.Attempt.TenantID, write.Attempt.ExportID, write.Attempt.JobID,
				write.Attempt.AttemptID, write.Attempt.Generation,
				write.Attempt.LeaseOwner, write.EffectID,
				write.EffectRequestSHA256[:], write.ObservedAt).Scan(&persisted)
		})
	if err != nil {
		return mapDispatchError(err)
	}
	if !persisted {
		return exportapp.ErrUnavailable
	}
	return nil
}

func (storage *WorkerStore) FinishRender(ctx context.Context,
	write exportapp.RenderFailure,
) (exportapp.State, error) {
	if !validFailureWrite(write) {
		return "", exportapp.ErrInvalid
	}
	var state exportapp.State
	err := storage.inTenant(ctx, write.Attempt.TenantID,
		func(boundContext context.Context, bound *workerBound) error {
			schema, err := workerSchema(boundContext, bound)
			if err != nil {
				return err
			}
			return bound.queryRow(boundContext, workerFailSQL, schema,
				write.Attempt.TenantID, write.Attempt.ExportID, write.Attempt.JobID,
				write.Attempt.AttemptID, write.Attempt.Generation,
				write.Attempt.LeaseOwner, nullableUUID(write.EffectID),
				nullableDigest(write.EffectRequestSHA256), write.FailureClass,
				write.Code, write.FailureSHA256[:], write.Retryable,
				nullableString(write.ProviderReference), write.FinishedAt).Scan(&state)
		})
	if err != nil {
		return "", mapDispatchError(err)
	}
	if state != exportapp.StateQueued && state != exportapp.StateFailed &&
		state != exportapp.StateCancelled {
		return "", exportapp.ErrUnavailable
	}
	return state, nil
}

func validFailureWrite(write exportapp.RenderFailure) bool {
	effectPair := write.EffectID == uuid.Nil && write.EffectRequestSHA256 == [32]byte{} ||
		write.EffectID.Version() == 7 && write.EffectRequestSHA256 != [32]byte{}
	return validAttemptIdentity(write.Attempt) && effectPair &&
		(write.FailureClass == "transient" || write.FailureClass == "permanent" ||
			write.FailureClass == "conflict" || write.FailureClass == "cancelled") &&
		failureCodePattern.MatchString(write.Code) && write.FailureSHA256 != [32]byte{} &&
		(!write.Retryable || write.FailureClass == "transient" || write.FailureClass == "conflict") &&
		(write.FailureClass != "cancelled" || !write.Retryable) && !write.FinishedAt.IsZero()
}

func nullableUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}

func nullableDigest(value [32]byte) any {
	if value == [32]byte{} {
		return nil
	}
	return value[:]
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
