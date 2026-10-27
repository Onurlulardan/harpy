package documents

import (
	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func parseArchiveExpiryClaim(claim platformasync.DatabaseJobClaim) (
	jobcontract.Reference, foundation.ID, error,
) {
	if claim == nil || claim.Kind() != asyncmanifestv1.DocumentsArchiveExpireJobKind ||
		claim.PayloadRevision() != 1 {
		return jobcontract.Reference{}, foundation.ID{}, foundation.ErrInvalidID
	}
	tenant, tenantScoped := claim.TenantID()
	parsed, err := jobcontract.ParseArchive(claim.PayloadReference())
	if err != nil || !tenantScoped || parsed.TenantID != jobcontract.ID(tenant) {
		return jobcontract.Reference{}, foundation.ID{}, foundation.ErrInvalidID
	}
	_, digest, err := jobcontract.Archive(parsed.TenantID, parsed.ResourceID)
	if err != nil || digest != claim.PayloadSHA256() {
		return jobcontract.Reference{}, foundation.ID{}, foundation.ErrInvalidID
	}
	return parsed, foundation.ID(tenant), nil
}
