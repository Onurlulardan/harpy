package work

import (
	"encoding/hex"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/api/jobcontract"
	jobapp "github.com/Onurlulardan/harpy/core/internal/modules/work/application/jobs"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func parseWorkJobClaim(claim platformasync.DatabaseJobClaim,
	now time.Time,
) (jobapp.Claim, error) {
	if claim == nil || claim.JobID() == uuid.Nil ||
		claim.JobID().Version() != uuid.Version(7) || claim.PayloadRevision() != 1 ||
		claim.LeaseGeneration() < 1 || now.IsZero() || now.Location() != time.UTC {
		return jobapp.Claim{}, ports.ErrInvalid
	}
	tenantID, tenantScoped := claim.TenantID()
	parsed, err := jobcontract.Parse(claim.Kind(), claim.PayloadReference())
	if err != nil || !tenantScoped || tenantID.String() != parsed.TenantID {
		return jobapp.Claim{}, ports.ErrInvalid
	}
	_, expectedDigest, err := jobcontract.Build(claim.Kind(), parsed.TenantID, parsed.TargetID)
	decoded, decodeErr := hex.DecodeString(claim.PayloadSHA256())
	if err != nil || decodeErr != nil || len(decoded) != 32 ||
		claim.PayloadSHA256() != expectedDigest {
		return jobapp.Claim{}, ports.ErrInvalid
	}
	var digest [32]byte
	copy(digest[:], decoded)
	targetID, targetErr := uuid.Parse(parsed.TargetID)
	if targetErr != nil {
		return jobapp.Claim{}, ports.ErrInvalid
	}
	result, err := jobapp.NewPlatformInput(claim.JobID(), tenantID, targetID,
		claim.Kind(), digest, claim.LeaseGeneration(), now)
	if err != nil {
		return jobapp.Claim{}, ports.ErrInvalid
	}
	return result, nil
}
