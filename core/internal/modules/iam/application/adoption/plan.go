package adoption

type CategoryEvidence struct {
	Category   string `json:"category"`
	RowCount   int    `json:"row_count"`
	RootSHA256 string `json:"root_sha256"`
}

type Plan struct {
	ContractVersion string             `json:"contract_version"`
	TenantID        string             `json:"tenant_id"`
	MappingSHA256   string             `json:"mapping_sha256"`
	SourceSHA256    string             `json:"source_sha256"`
	Categories      []CategoryEvidence `json:"categories"`
	Issues          []Issue            `json:"issues"`
	Eligible        bool               `json:"eligible"`
}

const PlanContractVersion = "harpy/iam-v1-adoption-plan/v1"
