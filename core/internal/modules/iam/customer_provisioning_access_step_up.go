package iam

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func customerProvisioningStepUp(
	claims accesstoken.Claims,
	purpose [32]byte,
	now time.Time,
) (time.Time, bool) {
	actual, present := claims.StepUpPurposeHash()
	authenticatedAt := time.Unix(claims.AuthenticationTimeUnix(), 0).UTC()
	return authenticatedAt, present && purpose != [32]byte{} && actual == purpose &&
		claims.Assurance() == accesstoken.AssuranceStepUp &&
		claims.AuthenticationTimeUnix() > 0 && !authenticatedAt.After(now) &&
		now.Sub(authenticatedAt) <= 10*time.Minute
}
