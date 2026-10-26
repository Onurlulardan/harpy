package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *PostRestoreHTTP) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodPost, Pattern: "/api/auth/post-restore/challenge",
			Handler: http.HandlerFunc(handler.challenge)},
		{Method: http.MethodPost, Pattern: "/api/auth/post-restore/reactivate-owner",
			Handler: http.HandlerFunc(handler.reactivateOwner)},
	}
}
