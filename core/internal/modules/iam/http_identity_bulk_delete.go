package iam

import (
	"net/http"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

type identityBulkDeleteDocument struct {
	IDs               []string         `json:"ids"`
	ExpectedRevisions map[string]int64 `json:"expected_revisions"`
}

type identityBulkDeleteAction func(
	contextClaims accesstoken.Claims, input IdentityBulkDelete,
) (int, error)

func (handler *IdentityHTTP) deleteUsers(w http.ResponseWriter, r *http.Request) {
	handler.deleteIdentityCollection(w, r, func(
		claims accesstoken.Claims, input IdentityBulkDelete,
	) (int, error) {
		return handler.application.DeleteUsers(r.Context(), claims, input)
	})
}

func (handler *IdentityHTTP) deleteRoles(w http.ResponseWriter, r *http.Request) {
	handler.deleteIdentityCollection(w, r, func(
		claims accesstoken.Claims, input IdentityBulkDelete,
	) (int, error) {
		return handler.application.DeleteRoles(r.Context(), claims, input)
	})
}

func (handler *IdentityHTTP) deleteOrganizations(w http.ResponseWriter, r *http.Request) {
	handler.deleteIdentityCollection(w, r, func(
		claims accesstoken.Claims, input IdentityBulkDelete,
	) (int, error) {
		return handler.application.DeleteOrganizations(r.Context(), claims, input)
	})
}

func (handler *IdentityHTTP) deleteIdentityCollection(
	w http.ResponseWriter, r *http.Request, action identityBulkDeleteAction,
) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailureAuthentication)
		return
	}
	key, valid := identityIdempotencyKey(r)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailureInvalid)
		return
	}
	document, err := decodeIdentityDocument[identityBulkDeleteDocument](w, r)
	if err != nil {
		handler.writeIdentityInputFailure(w, r, err)
		return
	}
	input, valid := identityBulkDeleteInput(key, document)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailureInvalid)
		return
	}
	count, err := action(claims, input)
	if err != nil {
		handler.writeIdentityApplicationFailure(w, r, err)
		return
	}
	handler.writeDelete(w, r, count)
}

func identityBulkDeleteInput(
	key string, document identityBulkDeleteDocument,
) (IdentityBulkDelete, bool) {
	if key == "" || len(document.IDs) < 1 ||
		len(document.IDs) > 100 || len(document.ExpectedRevisions) != len(document.IDs) {
		return IdentityBulkDelete{}, false
	}
	items := make([]IdentityMutation, 0, len(document.IDs))
	seen := make(map[string]struct{}, len(document.IDs))
	for _, rawID := range document.IDs {
		id, valid := decodeUUID(rawID)
		revision, found := document.ExpectedRevisions[rawID]
		if !valid || !found || revision < 1 {
			return IdentityBulkDelete{}, false
		}
		if _, duplicate := seen[rawID]; duplicate {
			return IdentityBulkDelete{}, false
		}
		seen[rawID] = struct{}{}
		items = append(items, IdentityMutation{ID: id, ExpectedRevision: revision})
	}
	sort.Slice(items, func(left, right int) bool {
		return items[left].ID.String() < items[right].ID.String()
	})
	return IdentityBulkDelete{IdempotencyKey: key, Items: items}, true
}
