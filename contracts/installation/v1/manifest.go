package installationv1

const (
	ContractVersion = "installation-manifest.v1"
	MaximumBytes    = 32 * 1024
)

type DeploymentMode string

const (
	DeploymentOnPremise DeploymentMode = "onpremise"
	DeploymentSaaS      DeploymentMode = "saas"
)

type TenantMetadata struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type AdminMetadata struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type manifestWire struct {
	ContractVersion string         `json:"contract_version"`
	EnvironmentID   string         `json:"environment_id"`
	DeploymentMode  DeploymentMode `json:"deployment_mode"`
	PrimaryTenant   TenantMetadata `json:"primary_tenant"`
	FirstAdmin      AdminMetadata  `json:"first_admin"`
	RequestSHA256   string         `json:"request_sha256"`
}

type Manifest struct {
	wire      manifestWire
	envID     [16]byte
	canonical []byte
	sha256    string
}

func (manifest Manifest) EnvironmentID() [16]byte { return manifest.envID }
func (manifest Manifest) EnvironmentIDString() string {
	return manifest.wire.EnvironmentID
}
func (manifest Manifest) DeploymentMode() DeploymentMode {
	return manifest.wire.DeploymentMode
}
func (manifest Manifest) PrimaryTenant() TenantMetadata { return manifest.wire.PrimaryTenant }
func (manifest Manifest) FirstAdmin() AdminMetadata     { return manifest.wire.FirstAdmin }
func (manifest Manifest) RequestSHA256() string         { return manifest.wire.RequestSHA256 }
func (manifest Manifest) SHA256() string                { return manifest.sha256 }
func (manifest Manifest) CanonicalBytes() []byte {
	return append([]byte(nil), manifest.canonical...)
}
