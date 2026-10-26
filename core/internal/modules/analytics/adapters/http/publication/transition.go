package publicationhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
)

func (handler *Handler) transition(
	writer http.ResponseWriter, request *http.Request, operation string,
) (httptransport.Access, publicationapp.TransitionRequest, error) {
	permission := httptransport.PermissionPublicationRead
	if operation == httptransport.OperationPublicationRevoke {
		permission = httptransport.PermissionPublicationRevoke
	}
	access, err := handler.runtime.AccessFor(request, permission, operation)
	if err != nil {
		return httptransport.Access{}, publicationapp.TransitionRequest{}, err
	}
	releaseID, idOK := httptransport.PathV7(request, "release_id")
	key, keyOK := httptransport.IdempotencyKey(request)
	if !idOK || !keyOK || request.URL.RawQuery != "" {
		return httptransport.Access{}, publicationapp.TransitionRequest{}, httptransport.ErrInvalid
	}
	document, err := httptransport.Decode[transitionDocument](writer, request)
	if err != nil {
		return httptransport.Access{}, publicationapp.TransitionRequest{}, err
	}
	return access, publicationapp.TransitionRequest{ReleaseID: releaseID,
		ExpectedRevision: document.ExpectedRevision, IdempotencyKey: key}, nil
}
