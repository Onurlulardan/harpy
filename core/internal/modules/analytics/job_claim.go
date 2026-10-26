package analytics

import (
	"crypto/sha256"
	"fmt"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func validateAnalyticsJobClaim(
	claim platformasync.DatabaseJobClaim, expectedKind string,
) error {
	if claim == nil || claim.Kind() != expectedKind || claim.PayloadRevision() != 1 ||
		claim.LeaseGeneration() < 1 || !validJobUUID(claim.JobID()) ||
		!validJobUUID(claim.AttemptID()) {
		return fmt.Errorf("Analytics job claim is invalid")
	}
	tenantID, tenantScoped := claim.TenantID()
	if !tenantScoped || tenantID == uuid.Nil {
		return fmt.Errorf("Analytics job tenant is invalid")
	}
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(claim.PayloadReference())))
	if claim.PayloadSHA256() != digest {
		return fmt.Errorf("Analytics job payload digest differs")
	}
	return nil
}

func validJobUUID(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7) &&
		value.Variant() == uuid.RFC4122
}
