package licensev2

type IssueRequest struct {
	Schema                 string             `json:"schema"`
	OperationID            string             `json:"operation_id"`
	LicenseID              string             `json:"license_id"`
	LicenseeOrganizationID string             `json:"licensee_organization_id"`
	EnvironmentID          string             `json:"environment_id"`
	DeploymentMode         string             `json:"deployment_mode"`
	DeliveryKind           DeliveryKind       `json:"delivery_kind"`
	Nonce                  string             `json:"nonce"`
	CurrentSequence        int64              `json:"current_sequence"`
	CurrentDigest          *string            `json:"current_digest"`
	ComponentVersions      []ComponentVersion `json:"component_versions"`
}

type ComponentVersion struct {
	Component string `json:"component"`
	Version   string `json:"version"`
}
