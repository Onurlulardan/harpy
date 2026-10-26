package installationv1

import (
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

type BuildInput struct {
	EnvironmentID  string
	DeploymentMode DeploymentMode
	PrimaryTenant  TenantMetadata
	FirstAdmin     AdminMetadata
}

func Build(input BuildInput) (Manifest, error) {
	wire := manifestWire{
		ContractVersion: ContractVersion,
		EnvironmentID:   input.EnvironmentID,
		DeploymentMode:  input.DeploymentMode,
		PrimaryTenant:   input.PrimaryTenant,
		FirstAdmin:      input.FirstAdmin,
	}
	wire.RequestSHA256 = requestDigest(wire)
	if !validWire(wire) {
		return Manifest{}, fmt.Errorf("installation manifest is invalid")
	}
	raw, err := json.Marshal(wire)
	if err != nil {
		return Manifest{}, fmt.Errorf("installation manifest is invalid")
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(raw, MaximumBytes)
	if err != nil {
		return Manifest{}, fmt.Errorf("installation manifest is invalid")
	}
	return Parse(canonical)
}
