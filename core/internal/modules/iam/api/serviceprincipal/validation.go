package serviceprincipal

import (
	"errors"
	"regexp"
	"time"
)

var ErrInvalidVerification = errors.New("service principal verification is invalid")

var capabilityPattern = regexp.MustCompile(
	`^[a-z][a-z0-9]*([._-][a-z0-9]+)*$`,
)

var resourcePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

func (request VerificationRequest) Validate() error {
	if !stable(request.PrincipalID) || !stable(request.GrantID) ||
		request.GrantRevision < 1 || !validCapability(request.Capability) ||
		request.ManifestSHA256 == [32]byte{} || !canonicalTime(request.VerifyAt) ||
		!validScope(request.Scope) {
		return ErrInvalidVerification
	}
	return nil
}

func (evidence VerificationEvidence) Validate() error {
	request := VerificationRequest{
		PrincipalID: evidence.PrincipalID, GrantID: evidence.GrantID,
		GrantRevision: evidence.GrantRevision, Capability: evidence.Capability,
		Scope: evidence.Scope, ManifestSHA256: evidence.ManifestSHA256,
		VerifyAt: evidence.VerifiedAt,
	}
	if request.Validate() != nil || evidence.PrincipalRevision < 1 ||
		evidence.AuthorizationVersion < 1 {
		return ErrInvalidVerification
	}
	return nil
}

func (evidence VerificationEvidence) Matches(request VerificationRequest) bool {
	return request.Validate() == nil && evidence.Validate() == nil &&
		evidence.PrincipalID == request.PrincipalID && evidence.GrantID == request.GrantID &&
		evidence.GrantRevision == request.GrantRevision &&
		evidence.Capability == request.Capability && evidence.Scope == request.Scope &&
		evidence.ManifestSHA256 == request.ManifestSHA256 &&
		evidence.VerifiedAt.Equal(request.VerifyAt)
}

func validScope(scope Scope) bool {
	if !resourcePattern.MatchString(scope.Owner) ||
		!resourcePattern.MatchString(scope.ResourceKind) {
		return false
	}
	switch scope.Kind {
	case "tenant":
		return scope.ResourceID == [16]byte{}
	case "resource":
		return stable(scope.ResourceID)
	default:
		return false
	}
}

func validCapability(value string) bool {
	return len(value) >= 3 && len(value) <= 100 && capabilityPattern.MatchString(value)
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}

func stable(value [16]byte) bool {
	return value != [16]byte{} && value[6]>>4 == 7 && value[8]&0xc0 == 0x80
}
