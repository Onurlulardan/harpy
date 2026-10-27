package licensev2

import "encoding/json"

type SnapshotEnvelope struct {
	Schema           string          `json:"schema"`
	KeyID            string          `json:"kid"`
	Algorithm        string          `json:"algorithm"`
	Canonicalization string          `json:"canonicalization"`
	PayloadSHA256    string          `json:"payload_sha256"`
	Payload          json.RawMessage `json:"payload"`
	Signature        string          `json:"signature"`
}

type SnapshotPayload struct {
	Schema                 string       `json:"schema"`
	Issuer                 string       `json:"issuer"`
	Audience               string       `json:"audience"`
	Product                string       `json:"product"`
	LicenseID              string       `json:"license_id"`
	LicenseeOrganizationID string       `json:"licensee_organization_id"`
	EnvironmentID          string       `json:"environment_id"`
	DeploymentMode         string       `json:"deployment_mode"`
	DeliveryKind           DeliveryKind `json:"delivery_kind"`
	Sequence               int64        `json:"sequence"`
	Status                 Status       `json:"status"`
	IssuedAt               string       `json:"issued_at"`
	NotBefore              string       `json:"not_before"`
	RefreshAfter           string       `json:"refresh_after"`
	GrantValidUntil        string       `json:"grant_valid_until"`
	GraceUntil             string       `json:"grace_until"`
	CommercialExpiresAt    string       `json:"commercial_expires_at"`
	Features               Features     `json:"features"`
	Limits                 Limits       `json:"limits"`
	ReleaseChannel         string       `json:"release_channel"`
}

type Features struct {
	FormBuilder      bool `json:"formBuilder"`
	WorkflowBuilder  bool `json:"workflowBuilder"`
	APIBuilder       bool `json:"apiBuilder"`
	DashboardBuilder bool `json:"dashboardBuilder"`
	AIAssistant      bool `json:"aiAssistant"`
}

type Limits struct {
	MaxUsers        int64 `json:"maxUsers"`
	MaxForms        int64 `json:"maxForms"`
	MaxWorkflows    int64 `json:"maxWorkflows"`
	MaxAPIEndpoints int64 `json:"maxApiEndpoints"`
	MaxDashboards   int64 `json:"maxDashboards"`
	MaxTenants      int64 `json:"maxTenants"`
}
