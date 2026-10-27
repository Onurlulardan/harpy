package queryhttp

import (
	"encoding/hex"
	"encoding/json"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
)

type view struct {
	QueryID         string          `json:"query_id"`
	VersionID       string          `json:"version_id"`
	VersionNumber   uint64          `json:"version_number"`
	RootRevision    uint64          `json:"root_revision"`
	ReferenceEpoch  uint64          `json:"reference_epoch"`
	VersionRevision uint64          `json:"version_revision"`
	Lifecycle       string          `json:"lifecycle"`
	VersionState    string          `json:"version_state"`
	CachePolicy     string          `json:"cache_policy"`
	Name            string          `json:"name"`
	Query           json.RawMessage `json:"query"`
	Parameters      json.RawMessage `json:"parameters"`
	Output          json.RawMessage `json:"output"`
	ManifestSHA256  string          `json:"manifest_sha256"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

func present(value queryapp.Snapshot) view {
	return view{QueryID: value.QueryID.String(), VersionID: value.VersionID.String(),
		VersionNumber: value.VersionNumber, RootRevision: value.RootRevision,
		ReferenceEpoch: value.ReferenceEpoch, VersionRevision: value.VersionRevision,
		Lifecycle: string(value.Lifecycle), VersionState: string(value.VersionState),
		CachePolicy: string(value.CachePolicy), Name: value.Name,
		Query: json.RawMessage(value.QueryDocument), Parameters: json.RawMessage(value.ParameterDocument),
		Output: json.RawMessage(value.OutputDocument), ManifestSHA256: hex.EncodeToString(value.ManifestSHA256[:]),
		CreatedAt: value.CreatedAt.Format(httptransport.MicrosecondTime),
		UpdatedAt: value.UpdatedAt.Format(httptransport.MicrosecondTime)}
}

func presentPage(value queryapp.Page) map[string]any {
	items := make([]view, len(value.Items))
	for index := range value.Items {
		items[index] = present(value.Items[index])
	}
	var after any
	if value.After != nil {
		after = value.After.String()
	}
	return map[string]any{"items": items, "after": after}
}
