package licenseauthoritysource

import (
	"regexp"
	"strings"

	licenseauthority "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
	"github.com/google/uuid"
)

var (
	componentPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]{0,63}$`)
	digestPattern    = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

func validConnectedRequest(request licenseauthority.IssueRequest) bool {
	if request.OperationID == uuid.Nil || request.Binding.LicenseID == uuid.Nil ||
		request.Binding.LicenseeOrganizationID == uuid.Nil ||
		request.Binding.EnvironmentID == uuid.Nil ||
		request.Binding.DeliveryKind != licenseauthority.DeliveryConnected ||
		request.Nonce == [32]byte{} || request.CurrentSequence < 0 ||
		len(request.ComponentVersions) == 0 || len(request.ComponentVersions) > 64 {
		return false
	}
	if (request.CurrentSequence == 0) != (request.CurrentDigest == "") ||
		(request.CurrentDigest != "" && !digestPattern.MatchString(request.CurrentDigest)) {
		return false
	}
	previous := ""
	for _, version := range request.ComponentVersions {
		if !componentPattern.MatchString(version.Component) ||
			!validVersion(version.Version) || version.Component <= previous {
			return false
		}
		previous = version.Component
	}
	return true
}

func validVersion(value string) bool {
	return value != "" && len(value) <= 128 && strings.TrimSpace(value) == value &&
		!strings.ContainsAny(value, "\x00\r\n")
}
