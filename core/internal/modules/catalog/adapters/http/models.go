package cataloghttp

import "time"

type Package struct {
	PackageID string    `json:"package_id"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type PackageVersion struct {
	PackageID      string    `json:"package_id"`
	Slug           string    `json:"slug"`
	Version        string    `json:"version"`
	ManifestSHA256 string    `json:"manifest_sha256"`
	KeyID          string    `json:"key_id"`
	RegisteredAt   time.Time `json:"registered_at"`
}

type PackageDependency struct {
	PackageID    string `json:"package_id"`
	VersionRange string `json:"version_range"`
	Kind         string `json:"kind"`
}

type Installation struct {
	InstallationID  string  `json:"installation_id"`
	PackageID       string  `json:"package_id"`
	ActiveVersion   *string `json:"active_version"`
	State           string  `json:"state"`
	ReferenceEpoch  uint64  `json:"reference_epoch"`
	Revision        uint64  `json:"revision"`
	LastOperationID *string `json:"last_operation_id"`
}

type Operation struct {
	OperationID string    `json:"operation_id"`
	Kind        string    `json:"kind"`
	State       string    `json:"state"`
	PlanSHA256  string    `json:"plan_sha256"`
	StartedAt   time.Time `json:"started_at"`
	FailureCode string    `json:"failure_code"`
	Revision    uint64    `json:"revision"`
	Replayed    bool      `json:"replayed"`
}

type PurgeApproval struct {
	ApprovalReceiptID string    `json:"approval_receipt_id"`
	OperationID       string    `json:"operation_id"`
	RequestedBy       string    `json:"requested_by"`
	ApprovedBy        string    `json:"approved_by"`
	ApprovalSHA256    string    `json:"approval_sha256"`
	ApprovedAt        time.Time `json:"approved_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	Replayed          bool      `json:"replayed"`
}

type DemoComponentStatus struct {
	ComponentID string `json:"component_id"`
	SourceRunID string `json:"source_run_id"`
	RecordCount int    `json:"record_count"`
}

type DemoStatus struct {
	InstallationID string                `json:"installation_id"`
	State          string                `json:"state"`
	Components     []DemoComponentStatus `json:"components"`
	UpdatedAt      *time.Time            `json:"updated_at"`
}

type DemoRemovalComponent struct {
	ComponentID string   `json:"component_id"`
	SourceRunID string   `json:"source_run_id"`
	RecordIDs   []string `json:"record_ids"`
}

type DemoRemovalPlan struct {
	InstallationID string                 `json:"installation_id"`
	Components     []DemoRemovalComponent `json:"components"`
	RecordCount    int                    `json:"record_count"`
	PlanSHA256     string                 `json:"plan_sha256"`
}

type MenuItem struct {
	ItemID    string  `json:"item_id"`
	PackageID string  `json:"package_id"`
	ParentID  *string `json:"parent_id"`
	Order     int     `json:"order"`
	Kind      string  `json:"kind"`
	Label     string  `json:"label"`
	TargetID  *string `json:"target_id"`
	Visible   bool    `json:"visible"`
}

type Menu struct {
	VersionID string     `json:"version_id"`
	Number    uint64     `json:"number"`
	Revision  uint64     `json:"revision"`
	Items     []MenuItem `json:"items"`
}

type MenuOverrideReceipt struct {
	ItemID   string `json:"item_id"`
	Revision uint64 `json:"revision"`
	Replayed bool   `json:"replayed"`
}
