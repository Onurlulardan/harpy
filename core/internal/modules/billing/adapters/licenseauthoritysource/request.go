package licenseauthoritysource

import (
	"encoding/base64"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	licensev2 "github.com/Onurlulardan/harpy/contracts/license/v2"
	licenseauthority "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
)

func marshalIssueRequest(request licenseauthority.IssueRequest) ([]byte, error) {
	if !validConnectedRequest(request) {
		return nil, ErrInvalidConfiguration
	}
	versions := make([]licensev2.ComponentVersion, len(request.ComponentVersions))
	for index, version := range request.ComponentVersions {
		versions[index] = licensev2.ComponentVersion{
			Component: version.Component, Version: version.Version,
		}
	}
	var digest *string
	if request.CurrentDigest != "" {
		value := request.CurrentDigest
		digest = &value
	}
	wire := licensev2.IssueRequest{
		Schema: licensev2.IssueRequestSchema, OperationID: request.OperationID.String(),
		LicenseID:              request.Binding.LicenseID.String(),
		LicenseeOrganizationID: request.Binding.LicenseeOrganizationID.String(),
		EnvironmentID:          request.Binding.EnvironmentID.String(), DeploymentMode: "onpremise",
		DeliveryKind:    request.Binding.DeliveryKind,
		Nonce:           base64.RawURLEncoding.EncodeToString(request.Nonce[:]),
		CurrentSequence: request.CurrentSequence, CurrentDigest: digest,
		ComponentVersions: versions,
	}
	document, err := json.Marshal(wire)
	if err != nil {
		return nil, err
	}
	return canonicaljsonv1.CanonicalizeObject(
		document, licenseauthority.MaximumDocumentBytes,
	)
}
