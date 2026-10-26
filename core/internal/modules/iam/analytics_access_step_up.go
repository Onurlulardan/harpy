package iam

import (
	"crypto/sha256"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func analyticsStepUpPurpose(tenantID uuid.UUID, operation string) [sha256.Size]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy/analytics/step-up/v1\x00"))
	_, _ = hash.Write(tenantID[:])
	_, _ = hash.Write([]byte(operation))
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func analyticsStepUpTime(
	claims accesstoken.Claims, operation string, now time.Time,
) (time.Time, bool) {
	if !analyticsOperationRequiresStepUp(operation) {
		return time.Time{}, true
	}
	purpose, found := claims.StepUpPurposeHash()
	authenticatedAt := time.Unix(claims.AuthenticationTimeUnix(), 0).UTC()
	return authenticatedAt, found && claims.Assurance() == accesstoken.AssuranceStepUp &&
		purpose == analyticsStepUpPurpose(claims.Tenant(), operation) &&
		claims.AuthenticationTimeUnix() > 0 && !authenticatedAt.After(now) &&
		now.Sub(authenticatedAt) <= 10*time.Minute
}
