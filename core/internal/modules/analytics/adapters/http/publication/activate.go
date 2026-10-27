package publicationhttp

import (
	"encoding/base64"
	"net/http"
	"time"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
)

func (handler *Handler) Activate(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionPublicationRead, httptransport.OperationPublicationActivate)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	releaseID, idOK := httptransport.PathV7(request, "release_id")
	key, keyOK := httptransport.IdempotencyKey(request)
	if !idOK || !keyOK || request.URL.RawQuery != "" {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	document, err := httptransport.Decode[activateDocument](writer, request)
	if err != nil || document.CapabilityTTLSeconds > 30*24*60*60 {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	result, err := handler.application.ActivatePublication(request.Context(), access, publicationapp.ActivateRequest{
		ReleaseID: releaseID, ExpectedRevision: document.ExpectedRevision,
		CapabilityTTL: time.Duration(document.CapabilityTTLSeconds) * time.Second, IdempotencyKey: key})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	response := map[string]any{"publication": present(result.Snapshot), "replay": result.Replay}
	if len(result.CapabilityToken) != 0 {
		response["capability_token"] = base64.RawURLEncoding.EncodeToString(result.CapabilityToken)
		defer clear(result.CapabilityToken)
	}
	handler.runtime.JSON(writer, request, http.StatusOK, response)
}
