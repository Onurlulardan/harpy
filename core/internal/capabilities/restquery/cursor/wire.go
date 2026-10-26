package cursor

import "encoding/json"

type wireClaims struct {
	Version                      string          `json:"version"`
	IssuedAt                     int64           `json:"issued_at"`
	ExpiresAt                    int64           `json:"expires_at"`
	TenantID                     string          `json:"tenant_id"`
	ScopeKind                    string          `json:"scope_kind"`
	ScopeID                      string          `json:"scope_id"`
	SubjectID                    string          `json:"subject_id"`
	Audience                     string          `json:"audience"`
	OperationID                  string          `json:"operation_id"`
	SnapshotAt                   string          `json:"snapshot_at"`
	AuthorizationContextSHA256   string          `json:"authorization_context_sha256"`
	ResourceOwner                string          `json:"resource_owner"`
	ResourceID                   string          `json:"resource_id"`
	PublishedVersionID           string          `json:"published_version_id"`
	CatalogSHA256                string          `json:"catalog_sha256"`
	QueryIdentitySHA256          string          `json:"query_identity_sha256"`
	MandatoryPredicateHMACSHA256 string          `json:"mandatory_predicate_hmac_sha256"`
	PageSize                     int             `json:"page_size"`
	EffectiveSort                []wireSort      `json:"effective_sort"`
	LastValues                   []wireLastValue `json:"last_values"`
}

type wireSort struct {
	FieldID     string `json:"field_id"`
	Direction   string `json:"direction"`
	LogicalType string `json:"logical_type"`
}

type wireLastValue struct {
	Null  bool            `json:"null"`
	Value json.RawMessage `json:"value" forgecrudNull:"allowed"`
}
