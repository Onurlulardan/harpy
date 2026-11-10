package postgres

import (
	"context"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type CredentialEventAuthority struct {
	TenantID, SubjectID uuid.UUID
	GroupGeneration     int64
	IAMGeneration       int64
	OperationID         string
}

type CredentialEventSink interface {
	Enqueue(context.Context, eventsv1.Envelope) error
	EnqueueJob(context.Context, jobsv1.Enqueue) (string, error)
}

type CredentialEventFactory func(any, CredentialEventAuthority) (CredentialEventSink, error)

type credentialEventBinding struct {
	sink CredentialEventSink
	err  error
}

func BindCredentialEvent(
	scope platformtx.CredentialTenantScope,
	subjectID uuid.UUID,
	factory CredentialEventFactory,
) (CredentialEventSink, error) {
	authority, err := platformtx.MintIAMCredentialEventAuthority(scope, subjectID)
	if err != nil || factory == nil {
		return nil, err
	}
	evidence := CredentialEventAuthority{
		TenantID: authority.ScopeID(), SubjectID: subjectID,
		GroupGeneration: authority.GroupGeneration(),
		IAMGeneration:   authority.IAMGeneration(), OperationID: authority.OperationID(),
	}
	binder, err := platformpostgres.NewOwnerBinder(
		"iam", func(database platformpostgres.BoundExecutor) credentialEventBinding {
			sink, buildErr := factory(database, evidence)
			return credentialEventBinding{sink: sink, err: buildErr}
		},
	)
	if err != nil {
		return nil, err
	}
	bound, err := platformtx.BindCredential(scope, binder)
	if err != nil || bound.err != nil || bound.sink == nil {
		if err != nil {
			return nil, err
		}
		return nil, bound.err
	}
	return bound.sink, nil
}

func BindProfileReader(scope platformtx.TenantScope) (ProfileReader, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		profileStorageOwner,
		func(database platformpostgres.BoundExecutor) *profileReader {
			return &profileReader{
				queryRow: func(
					ctx context.Context,
					query string,
					arguments ...any,
				) profileRow {
					return database.QueryRow(ctx, query, arguments...)
				},
				scopeCheck: func(ctx context.Context) bool {
					return profileScopeUsable(ctx, scope)
				},
			}
		},
	)
	if err != nil {
		return nil, errProfileSnapshotUnavailable
	}
	reader, err := platformtx.Bind(scope, binder)
	if err != nil || reader == nil {
		return nil, errProfileSnapshotUnavailable
	}
	return reader, nil
}

func profileScopeUsable(ctx context.Context, scope platformtx.TenantScope) bool {
	if ctx == nil || ctx.Err() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) {
		return false
	}
	return platformtx.RequireActive(scope, platformtx.Command) == nil ||
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) == nil
}
