package v1adoption

import (
	"time"

	"github.com/google/uuid"
)

type DeploymentMode string

const (
	ModeOnPremise DeploymentMode = "onpremise"
	ModeSaaS      DeploymentMode = "saas"
)

type TenantEvidence struct {
	TenantID         uuid.UUID `json:"tenant_id"`
	Status           string    `json:"status"`
	ControlCandidate bool      `json:"control_candidate"`
}

type ProtectedAccountEvidence struct {
	TenantID             uuid.UUID `json:"tenant_id"`
	UserID               uuid.UUID `json:"user_id"`
	Active               bool      `json:"active"`
	CanonicalIAMMapped   bool      `json:"canonical_iam_mapped"`
	ProtectedAdminBound  bool      `json:"protected_admin_bound"`
	UsableCredential     bool      `json:"usable_credential"`
	AuthorizationVersion int64     `json:"authorization_version"`
	MappingSHA256        string    `json:"mapping_sha256"`
}

type BillingKind string

const (
	BillingLicenseDerived BillingKind = "license_derived"
	BillingSaaSControl    BillingKind = "saas_control"
	BillingSaaSFree       BillingKind = "saas_free"
	BillingSaaSPaid       BillingKind = "saas_paid"
)

type OfferingPin struct {
	Key            string    `json:"key"`
	VersionID      uuid.UUID `json:"version_id"`
	Revision       int64     `json:"revision"`
	SHA256         string    `json:"sha256"`
	EffectiveAt    time.Time `json:"effective_at"`
	SourceSequence int64     `json:"source_sequence"`
	SourceSHA256   string    `json:"source_sha256"`
}

type BillingEvidence struct {
	TenantID                   uuid.UUID   `json:"tenant_id"`
	Kind                       BillingKind `json:"kind"`
	VerifiedV2Publication      bool        `json:"verified_v2_publication"`
	SubscriptionEvidenceSHA256 string      `json:"subscription_evidence_sha256"`
	Offering                   OfferingPin `json:"offering"`
}

type Inventory struct {
	EnvironmentID uuid.UUID                  `json:"environment_id"`
	Mode          DeploymentMode             `json:"mode"`
	Tenants       []TenantEvidence           `json:"tenants"`
	Protected     []ProtectedAccountEvidence `json:"protected_accounts"`
	Billing       []BillingEvidence          `json:"billing"`
}

type TenantBinding struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Purpose  string    `json:"purpose"`
}

type BaseAssignment struct {
	TenantID                   uuid.UUID   `json:"tenant_id"`
	Kind                       BillingKind `json:"kind"`
	Offering                   OfferingPin `json:"offering"`
	SubscriptionEvidenceSHA256 string      `json:"subscription_evidence_sha256"`
}

type Plan struct {
	EnvironmentID             uuid.UUID        `json:"environment_id"`
	Mode                      DeploymentMode   `json:"mode"`
	PrimaryTenantID           uuid.UUID        `json:"primary_tenant_id"`
	OwnerUserID               uuid.UUID        `json:"owner_user_id"`
	OwnerAuthorizationVersion int64            `json:"owner_authorization_version"`
	OwnerMappingSHA           string           `json:"owner_mapping_sha256"`
	TenantBindings            []TenantBinding  `json:"tenant_bindings"`
	BaseAssignments           []BaseAssignment `json:"base_assignments"`
	InventorySHA256           string           `json:"inventory_sha256"`
	PlanSHA256                string           `json:"plan_sha256"`
}
