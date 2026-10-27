package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *InvitationHTTP) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodPost, Pattern: "/api/invitations", Handler: http.HandlerFunc(handler.create)},
		{Method: http.MethodGet, Pattern: "/api/invitations", Handler: http.HandlerFunc(handler.list)},
		{Method: http.MethodGet, Pattern: "/api/invitations/verify", Handler: http.HandlerFunc(handler.verify)},
		{Method: http.MethodPost, Pattern: "/api/invitations/accept", Handler: http.HandlerFunc(handler.accept)},
		{Method: http.MethodGet, Pattern: "/api/invitations/{id}", Handler: http.HandlerFunc(handler.read)},
		{Method: http.MethodDelete, Pattern: "/api/invitations/{id}", Handler: http.HandlerFunc(handler.revoke)},
	}
}
