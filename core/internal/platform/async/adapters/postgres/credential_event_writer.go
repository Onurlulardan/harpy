package postgres

import (
	"context"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

type EventWriter interface {
	EnqueueEvent(context.Context, eventsv1.Envelope) (EnqueueResult, error)
	EnqueueJob(context.Context, jobsv1.Enqueue) (EnqueueResult, error)
}

type iamCredentialEventWriter struct {
	writer    *repository
	subjectID uuid.UUID
}

func NewIAMCredentialEventWriter(
	database any,
	tenantID uuid.UUID,
	groupGeneration int64,
	iamGeneration int64,
	subjectID uuid.UUID,
	operationID string,
	jobs jobsv1.Catalog,
	producers eventsv1.ProducerCatalog,
) (EventWriter, error) {
	bound, ok := database.(boundDatabase)
	stamp := authorityStamp{
		scopeID: tenantID, groupID: "iam_system_resources",
		groupGeneration: groupGeneration, domain: "iam", generation: iamGeneration,
	}
	if !ok || bound == nil || tenantID == uuid.Nil || groupGeneration < 1 ||
		iamGeneration < 1 || subjectID == uuid.Nil || operationID == "" {
		return nil, ErrAuthorityMismatch
	}
	return &iamCredentialEventWriter{
		writer: &repository{
			database: bound, authority: stamp, principalID: subjectID,
			operationID: operationID, jobs: jobs, producers: producers,
		},
		subjectID: subjectID,
	}, nil
}

func (writer *iamCredentialEventWriter) EnqueueJob(
	ctx context.Context, value jobsv1.Enqueue,
) (EnqueueResult, error) {
	if writer == nil || writer.writer == nil || writer.subjectID == uuid.Nil {
		return EnqueueResult{}, ErrAuthorityMismatch
	}
	return writer.writer.EnqueueJob(ctx, value)
}

func (writer *iamCredentialEventWriter) EnqueueEvent(
	ctx context.Context,
	value eventsv1.Envelope,
) (EnqueueResult, error) {
	if writer == nil || writer.writer == nil || writer.subjectID == uuid.Nil {
		return EnqueueResult{}, ErrAuthorityMismatch
	}
	actorID, err := uuid.Parse(value.ActorID)
	if err != nil || actorID != writer.subjectID {
		return EnqueueResult{}, ErrAuthorityMismatch
	}
	return writer.writer.EnqueueEvent(ctx, value)
}
