package queryhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
)

const (
	operationQueryPublish    = httptransport.OperationQueryPublish
	operationQueryRevalidate = httptransport.OperationQueryRevalidate
)

func (handler *Handler) publicationInput(
	writer http.ResponseWriter, request *http.Request, operation string,
) (httptransport.Access, queryapp.PublishRequest, error) {
	access, err := handler.runtime.AccessFor(
		request, httptransport.PermissionQueryPublish, operation,
	)
	if err != nil {
		return httptransport.Access{}, queryapp.PublishRequest{}, err
	}
	queryID, idOK := httptransport.PathV7(request, "query_id")
	key, keyOK := httptransport.IdempotencyKey(request)
	if !idOK || !keyOK || request.URL.RawQuery != "" {
		return httptransport.Access{}, queryapp.PublishRequest{}, httptransport.ErrInvalid
	}
	document, err := httptransport.Decode[publishDocument](writer, request)
	if err != nil {
		return httptransport.Access{}, queryapp.PublishRequest{}, err
	}
	draftID, draftOK := httptransport.WireV7(document.DraftVersionID)
	resolved, dependenciesOK := dependencies(document.Dependencies)
	if !draftOK || !dependenciesOK {
		return httptransport.Access{}, queryapp.PublishRequest{}, httptransport.ErrInvalid
	}
	return access, queryapp.PublishRequest{QueryID: queryID, DraftVersionID: draftID,
		ExpectedRootRevision:  document.ExpectedRootRevision,
		ExpectedDraftRevision: document.ExpectedDraftRevision, Dependencies: resolved,
		RevisionBoundCache: document.RevisionBoundCache, IdempotencyKey: key}, nil
}
