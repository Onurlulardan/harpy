package iam

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	serviceprincipal "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
	"github.com/google/uuid"
)

func workflowPublicationAuthorityStepUp(
	claims accesstoken.Claims,
	rawDefinitionID string,
	now time.Time,
) ([32]byte, time.Time, bool, bool, bool, bool) {
	definitionID, err := uuid.Parse(rawDefinitionID)
	if err != nil || definitionID.String() != rawDefinitionID ||
		definitionID.Version() != 7 || definitionID.Variant() != uuid.RFC4122 {
		return [32]byte{}, time.Time{}, false, false, false, false
	}
	expected, purposeErr := serviceprincipal.PublicationAuthorityStepUpPurpose(
		claims.Tenant(), definitionID,
	)
	purpose, found := claims.StepUpPurposeHash()
	authenticatedAt := time.Unix(claims.AuthenticationTimeUnix(), 0).UTC()
	password, webAuthn, totp := workflowAccessMethods(claims.Methods())
	valid := purposeErr == nil && found &&
		claims.Assurance() == accesstoken.AssuranceStepUp &&
		purpose == expected && claims.AuthenticationTimeUnix() > 0 &&
		!authenticatedAt.After(now) && now.Sub(authenticatedAt) <= 10*time.Minute &&
		(password || webAuthn || totp)
	return expected, authenticatedAt, password, webAuthn, totp, valid
}

func workflowAccessMethods(
	methods []accesstoken.AuthMethod,
) (bool, bool, bool) {
	var password, webAuthn, totp bool
	for _, method := range methods {
		switch method {
		case accesstoken.AuthMethodPassword:
			password = true
		case accesstoken.AuthMethodWebAuthn:
			webAuthn = true
		case accesstoken.AuthMethodTOTP:
			totp = true
		}
	}
	return password, webAuthn, totp
}
