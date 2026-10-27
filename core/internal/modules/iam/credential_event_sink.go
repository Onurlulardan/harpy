package iam

import (
	"context"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type credentialEventSink struct{ writer asyncpostgres.EventWriter }

func (sink credentialEventSink) Enqueue(
	ctx context.Context,
	value eventsv1.Envelope,
) error {
	_, err := sink.writer.EnqueueEvent(ctx, value)
	return err
}

func (sink credentialEventSink) EnqueueJob(
	ctx context.Context, value jobsv1.Enqueue,
) (string, error) {
	result, err := sink.writer.EnqueueJob(ctx, value)
	return result.ID, err
}

func BindCredentialEventSink(
	scope platformtx.CredentialTenantScope,
	subjectID uuid.UUID,
	catalogs platformasync.WriteCatalogs,
) (iampostgres.CredentialEventSink, error) {
	jobs, producers, err := catalogs.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("bind IAM credential event catalogs: %w", err)
	}
	return iampostgres.BindCredentialEvent(
		scope, subjectID,
		func(database any, authority iampostgres.CredentialEventAuthority) (
			iampostgres.CredentialEventSink, error,
		) {
			writer, bindErr := asyncpostgres.NewIAMCredentialEventWriter(
				database, authority.TenantID, authority.GroupGeneration,
				authority.IAMGeneration, authority.SubjectID,
				authority.OperationID, jobs, producers,
			)
			if bindErr != nil {
				return nil, bindErr
			}
			return credentialEventSink{writer: writer}, nil
		},
	)
}
