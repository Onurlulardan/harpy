package documents

import (
	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func parsePurgeClaim(claim platformasync.DatabaseJobClaim) (jobcontract.PurgeReference,
	foundation.ID, error,
) {
	if claim == nil || claim.Kind() != asyncmanifestv1.DocumentsItemPurgeJobKind ||
		claim.PayloadRevision() != 1 {
		return jobcontract.PurgeReference{}, foundation.ID{}, foundation.ErrInvalidID
	}
	tenant, tenantScoped := claim.TenantID()
	parsed, err := jobcontract.ParsePurge(claim.PayloadReference())
	if err != nil || !tenantScoped || parsed.TenantID != jobcontract.ID(tenant) {
		return jobcontract.PurgeReference{}, foundation.ID{}, foundation.ErrInvalidID
	}
	_, digest, err := jobcontract.PurgePage(parsed.TenantID, parsed.ResourceID,
		parsed.AfterDeletionID)
	if err != nil || digest != claim.PayloadSHA256() {
		return jobcontract.PurgeReference{}, foundation.ID{}, foundation.ErrInvalidID
	}
	return parsed, foundation.ID(tenant), nil
}
