package iam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type identityRecordEnvelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Message string `json:"message,omitempty"`
}

type identityDeleteEnvelope struct {
	Success      bool   `json:"success"`
	Message      string `json:"message,omitempty"`
	DeletedCount int    `json:"deleted_count"`
}

type identityQueryData struct {
	Items      []json.RawMessage       `json:"items"`
	Pagination identityQueryPagination `json:"pagination"`
}

type identityQueryPagination struct {
	HasMore       bool   `json:"has_more"`
	NextCursor    string `json:"next_cursor,omitempty"`
	NextTransport string `json:"next_transport,omitempty"`
	ExactCount    *int64 `json:"exact_count,omitempty"`
}

type identityQueryEnvelope struct {
	Success bool              `json:"success"`
	Data    identityQueryData `json:"data"`
}

func (handler *IdentityHTTP) writeRecord(
	w http.ResponseWriter, r *http.Request, status int, revision int64, record any,
) {
	if revision < 1 {
		handler.writeIdentityFailure(w, r, identityFailureUnavailable)
		return
	}
	flattened, err := identityFlattenPublicProfile(record)
	if err != nil {
		handler.report(err)
		handler.writeIdentityFailure(w, r, identityFailureUnavailable)
		return
	}
	w.Header().Set("ETag", fmt.Sprintf(`"revision-%d"`, revision))
	handler.writeJSON(w, r, status, identityRecordEnvelope{Success: true, Data: flattened})
}

func (handler *IdentityHTTP) writeDelete(w http.ResponseWriter, r *http.Request, count int) {
	handler.writeJSON(w, r, http.StatusOK, identityDeleteEnvelope{
		Success: true, DeletedCount: count,
	})
}

func (handler *IdentityHTTP) writeQuery(
	w http.ResponseWriter, r *http.Request, view IdentityQueryView,
) {
	items := make([]json.RawMessage, len(view.Items))
	for index, item := range view.Items {
		canonical, err := canonicaljsonv1.CanonicalizeObject(item, 64*1024)
		if err != nil || !identityQueryItemSafe(canonical) {
			clear(canonical)
			handler.report(fmt.Errorf("IAM query result item is invalid"))
			handler.writeIdentityFailure(w, r, identityFailureUnavailable)
			return
		}
		items[index] = canonical
	}
	response := identityQueryEnvelope{Success: true, Data: identityQueryData{
		Items: items, Pagination: identityQueryPagination{
			HasMore: view.HasMore, NextCursor: view.NextCursor,
			NextTransport: view.NextTransport, ExactCount: view.ExactCount,
		},
	}}
	handler.writeJSON(w, r, http.StatusOK, response)
}

func identityQueryItemSafe(document []byte) bool {
	var value map[string]json.RawMessage
	if json.Unmarshal(document, &value) != nil {
		return false
	}
	for _, key := range []string{
		"password", "password_hash", "email_normalized", "token_version",
		"authorization_version", "identifier_id",
	} {
		if _, exposed := value[key]; exposed {
			return false
		}
	}
	return !bytes.Equal(document, []byte("{}"))
}

func (handler *IdentityHTTP) writeJSON(
	w http.ResponseWriter, r *http.Request, status int, payload any,
) {
	if err := httpwire.WriteJSON(w, status, authHTTPCorrelationID(r), payload, false); err != nil {
		handler.report(fmt.Errorf("write IAM system-resource response: %w", err))
	}
}
