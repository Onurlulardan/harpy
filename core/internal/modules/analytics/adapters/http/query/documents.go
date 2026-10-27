package queryhttp

import "encoding/json"

type createDocument struct {
	Name       string          `json:"name"`
	Query      json.RawMessage `json:"query"`
	Parameters json.RawMessage `json:"parameters"`
	Output     json.RawMessage `json:"output"`
}

type updateDocument struct {
	ExpectedRootRevision    uint64          `json:"expected_root_revision"`
	ExpectedVersionRevision uint64          `json:"expected_version_revision"`
	Name                    string          `json:"name"`
	Query                   json.RawMessage `json:"query"`
	Parameters              json.RawMessage `json:"parameters"`
	Output                  json.RawMessage `json:"output"`
}

type dependencyDocument struct {
	OwnerDomain  string `json:"owner_domain"`
	ResourceKind string `json:"resource_kind"`
	ResourceID   string `json:"resource_id"`
	VersionID    string `json:"version_id"`
}

type publishDocument struct {
	DraftVersionID        string               `json:"draft_version_id"`
	ExpectedRootRevision  uint64               `json:"expected_root_revision"`
	ExpectedDraftRevision uint64               `json:"expected_draft_revision"`
	Dependencies          []dependencyDocument `json:"dependencies"`
	RevisionBoundCache    bool                 `json:"revision_bound_cache"`
}

type executeDocument struct {
	VersionID string          `json:"version_id"`
	Params    json.RawMessage `json:"params"`
}

type deleteDocument struct {
	ExpectedRootRevision uint64 `json:"expected_root_revision"`
}
