package iam

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/google/uuid"
)

type lifecycleSessionView struct {
	ID             uuid.UUID  `json:"id"`
	PortalBinding  *uuid.UUID `json:"portal_binding_id,omitempty"`
	AccessMode     string     `json:"access_mode"`
	DeviceKind     string     `json:"device_kind"`
	BrowserFamily  string     `json:"browser_family"`
	OSFamily       string     `json:"os_family"`
	Revision       int64      `json:"revision"`
	CreatedAt      time.Time  `json:"created_at"`
	LastSeenAt     time.Time  `json:"last_seen_at"`
	AbsoluteExpiry time.Time  `json:"absolute_expiry"`
}

type lifecycleSessionListResponse struct {
	Items []lifecycleSessionView `json:"items"`
}

func lifecycleSessionListView(items []lifecycle.Session) lifecycleSessionListResponse {
	views := make([]lifecycleSessionView, 0, len(items))
	for _, item := range items {
		view := lifecycleSessionView{
			ID: item.ID, AccessMode: item.AccessMode,
			DeviceKind: item.DeviceKind, BrowserFamily: item.BrowserFamily,
			OSFamily: item.OSFamily, Revision: item.Revision,
			CreatedAt: item.CreatedAt, LastSeenAt: item.LastSeenAt,
			AbsoluteExpiry: item.AbsoluteExpiry,
		}
		if item.PortalBindingID != uuid.Nil {
			portalID := item.PortalBindingID
			view.PortalBinding = &portalID
		}
		views = append(views, view)
	}
	return lifecycleSessionListResponse{Items: views}
}
