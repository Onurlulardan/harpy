package purge

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func buildMutation(ids foundation.IDSource, clock foundation.Clock,
	candidate Candidate,
) (Mutation, error) {
	operation, err := nextMutationID(ids)
	if err != nil {
		return Mutation{}, err
	}
	audit, err := nextMutationID(ids)
	if err != nil || audit == operation {
		return Mutation{}, ErrUnavailable
	}
	mutation := Mutation{Candidate: candidate, OperationID: operation,
		AuditID: audit, OccurredAt: clock.Now().UTC()}
	for _, version := range candidate.Versions {
		deletionID, idErr := nextMutationID(ids)
		if idErr != nil {
			return Mutation{}, idErr
		}
		mutation.Deletions = append(mutation.Deletions, Deletion{Version: version,
			DeletionID: deletionID, JobID: deletionID})
	}
	return mutation, nil
}

func nextMutationID(ids foundation.IDSource) (domain.ID, error) {
	value, err := ids.NewUUIDv7()
	if err != nil || !value.UUIDv7() {
		return domain.ID{}, ErrUnavailable
	}
	return value, nil
}
