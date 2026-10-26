package billing

import (
	"fmt"
	"time"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func providerFinalizeSuccessor(
	result providerapp.Finalization,
) (platformasync.DatabaseJobSuccessorInput, error) {
	jobID := result.FinalizerJobID
	if jobID == uuid.Nil || jobID.Version() != uuid.Version(7) {
		return platformasync.DatabaseJobSuccessorInput{}, providerapp.ErrConfiguration
	}
	return platformasync.DatabaseJobSuccessorInput{
		JobID: jobID, RunAt: result.OccurredAt.Add(time.Second),
		PayloadReference: "billing-provider-finalize/" + result.CommandID.String(),
		PayloadSHA256:    result.RequestSHA256,
		DedupeKey: fmt.Sprintf(
			"billing:provider:finalize:%s:%s", result.CommandID, jobID,
		),
		Priority: 100,
	}, nil
}
