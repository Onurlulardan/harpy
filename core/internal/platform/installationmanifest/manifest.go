package installationmanifest

import installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"

type Manifest = installationv1.Manifest
type BuildInput = installationv1.BuildInput
type TenantMetadata = installationv1.TenantMetadata
type AdminMetadata = installationv1.AdminMetadata
type DeploymentMode = installationv1.DeploymentMode

const (
	MaximumBytes        = installationv1.MaximumBytes
	DeploymentOnPremise = installationv1.DeploymentOnPremise
	DeploymentSaaS      = installationv1.DeploymentSaaS
)

func Parse(document []byte) (Manifest, error) {
	return installationv1.Parse(document)
}

func Build(input BuildInput) (Manifest, error) {
	return installationv1.Build(input)
}
