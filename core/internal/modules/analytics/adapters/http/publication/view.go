package publicationhttp

import (
	"encoding/base64"
	"encoding/hex"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"github.com/google/uuid"
)

type view struct {
	ReleaseID           string `json:"release_id"`
	ResourceKind        string `json:"resource_kind"`
	ResourceVersionID   string `json:"resource_version_id"`
	Audience            string `json:"audience"`
	State               string `json:"state"`
	CachePolicy         string `json:"cache_policy"`
	ManifestSHA256      string `json:"manifest_sha256"`
	RequesterID         string `json:"requester_id"`
	ApproverID          string `json:"approver_id,omitempty"`
	Revision            uint64 `json:"revision"`
	Generation          uint64 `json:"generation"`
	PublicID            any    `json:"public_id"`
	CapabilityExpiresAt any    `json:"capability_expires_at"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

func present(value publicationapp.Snapshot) view {
	result := view{ReleaseID: value.ReleaseID.String(), ResourceKind: string(value.ResourceKind),
		ResourceVersionID: value.ResourceVersionID.String(), Audience: string(value.Audience),
		State: string(value.State), CachePolicy: string(value.CachePolicy),
		ManifestSHA256: hex.EncodeToString(value.ManifestSHA256[:]), RequesterID: value.RequesterID.String(),
		Revision: value.Revision, Generation: value.Generation,
		CreatedAt: value.CreatedAt.Format(httptransport.MicrosecondTime),
		UpdatedAt: value.UpdatedAt.Format(httptransport.MicrosecondTime)}
	if value.ApproverID != uuid.Nil {
		result.ApproverID = value.ApproverID.String()
	}
	if value.PublicID != [16]byte{} {
		result.PublicID = base64.RawURLEncoding.EncodeToString(value.PublicID[:])
	}
	if value.CapabilityExpiresAt != nil {
		result.CapabilityExpiresAt = value.CapabilityExpiresAt.Format(httptransport.MicrosecondTime)
	}
	return result
}

func presentPage(value publicationapp.Page) map[string]any {
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
