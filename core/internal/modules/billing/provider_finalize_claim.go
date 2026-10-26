package billing

import (
	"fmt"
	"strings"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func providerFinalizeRequest(
	claim platformasync.DatabaseJobClaim,
) (providerapp.FinalizeCommandRequest, error) {
	tenantID, tenantScoped := claim.TenantID()
	prefix := "billing-provider-finalize/"
	if !tenantScoped || tenantID == uuid.Nil ||
		claim.Kind() != providerapp.ProviderFinalizeJobKind ||
		claim.PayloadRevision() != 1 ||
		!strings.HasPrefix(claim.PayloadReference(), prefix) ||
		len(claim.PayloadReference()) <= len(prefix) ||
		!providerPayloadDigest(claim.PayloadSHA256()) {
		return providerapp.FinalizeCommandRequest{},
			fmt.Errorf("billing provider finalizer claim is invalid")
	}
	commandID, err := uuid.Parse(strings.TrimPrefix(claim.PayloadReference(), prefix))
	if err != nil || commandID == uuid.Nil ||
		commandID.Version() != uuid.Version(7) ||
		claim.JobID() == uuid.Nil || claim.AttemptID() == uuid.Nil {
		return providerapp.FinalizeCommandRequest{},
			fmt.Errorf("billing provider finalizer claim is invalid")
	}
	return providerapp.FinalizeCommandRequest{
		TenantID: tenantID, JobID: claim.JobID(),
		AttemptID: claim.AttemptID(), CommandID: commandID,
		LeaseGeneration:  claim.LeaseGeneration(),
		PayloadRevision:  claim.PayloadRevision(),
		PayloadReference: claim.PayloadReference(),
		PayloadSHA256:    claim.PayloadSHA256(),
	}, nil
}

func providerPayloadDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') &&
			(character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}
