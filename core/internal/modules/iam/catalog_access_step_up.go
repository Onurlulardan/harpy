package iam

import (
	"crypto/sha256"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func catalogStepUpPurpose(
	tenantID uuid.UUID, operation, target string,
) [sha256.Size]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy/catalog/step-up/v1\x00"))
	_, _ = hash.Write(tenantID[:])
	_, _ = hash.Write([]byte(operation + "\x00" + target))
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func catalogStepUpTime(
	claims accesstoken.Claims, request CatalogAccessRequest, now time.Time,
) (time.Time, bool) {
	if !catalogOperationRequiresStepUp(request.Operation) {
		return time.Time{}, request.PurposeTarget == ""
	}
	purpose, found := claims.StepUpPurposeHash()
	authenticatedAt := time.Unix(claims.AuthenticationTimeUnix(), 0).UTC()
	return authenticatedAt, request.PurposeTarget != "" && found &&
		claims.Assurance() == accesstoken.AssuranceStepUp &&
		purpose == catalogStepUpPurpose(claims.Tenant(), request.Operation, request.PurposeTarget) &&
		claims.AuthenticationTimeUnix() > 0 && !authenticatedAt.After(now) &&
		now.Sub(authenticatedAt) <= 10*time.Minute
}
