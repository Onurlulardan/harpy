package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	layouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type layoutHTTPSuccessEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type layoutHTTPRevisionPage struct {
	Items []layoutHTTPRevisionView `json:"items"`
}

func (handler *systemResourceLayoutHTTPHandler) writeHTTPRevision(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	resource layouts.Resource,
	revision layouts.Revision,
	location string,
) {
	if location != "" {
		w.Header().Set("Location", location)
	}
	w.Header().Set("ETag", layoutHTTPETag(revision.RowRevision))
	handler.writeHTTPJSON(w, r, status, layoutHTTPSuccessEnvelope{
		Success: true, Data: layoutHTTPRevisionResponse(resource, revision),
	})
}

func (handler *systemResourceLayoutHTTPHandler) writeHTTPRevisionList(
	w http.ResponseWriter,
	r *http.Request,
	resource layouts.Resource,
	revisions []layouts.Revision,
) {
	items := make([]layoutHTTPRevisionView, len(revisions))
	for index, revision := range revisions {
		items[index] = layoutHTTPRevisionResponse(resource, revision)
	}
	handler.writeHTTPJSON(w, r, http.StatusOK, layoutHTTPSuccessEnvelope{
		Success: true, Data: layoutHTTPRevisionPage{Items: items},
	})
}

func (handler *systemResourceLayoutHTTPHandler) writeHTTPDelete(
	w http.ResponseWriter, r *http.Request, revisionID string,
) {
	handler.writeHTTPJSON(w, r, http.StatusOK, layoutHTTPSuccessEnvelope{
		Success: true, Data: map[string]any{
			"revision_id": revisionID, "deleted": true,
		},
	})
}

func (handler *systemResourceLayoutHTTPHandler) writeHTTPPresentation(
	w http.ResponseWriter, r *http.Request, document layouts.PresentationDocument,
) {
	encoded, err := json.Marshal(document)
	if err != nil {
		handler.writeHTTPFailure(w, r, layouts.ErrLayoutUnavailable)
		return
	}
	digest := sha256.Sum256(encoded)
	w.Header().Set("ETag", `"presentation-`+hex.EncodeToString(digest[:])+`"`)
	handler.writeHTTPJSON(w, r, http.StatusOK, layoutHTTPSuccessEnvelope{
		Success: true, Data: document,
	})
}

func (handler *systemResourceLayoutHTTPHandler) writeHTTPJSON(
	w http.ResponseWriter, r *http.Request, status int, payload any,
) {
	if err := httpwire.WriteJSON(
		w, status, layoutHTTPCorrelationID(r), payload, false,
	); err != nil {
		handler.report(errors.New("system resource layout HTTP response write failed"))
	}
}

func layoutHTTPETag(revision int64) string {
	return fmt.Sprintf(`"layout-row-revision-%d"`, revision)
}

func layoutHTTPRevisionLocation(resource layouts.Resource, id string) string {
	return fmt.Sprintf("/api/system-resources/%s/layout-revisions/%s", resource, id)
}

func layoutHTTPPublicationLocation(resource layouts.Resource, id string) string {
	return layoutHTTPRevisionLocation(resource, id) + "/publication"
}
