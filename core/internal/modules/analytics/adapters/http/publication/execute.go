package publicationhttp

import (
	"errors"
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
)

func (handler *Handler) Execute(writer http.ResponseWriter, request *http.Request) {
	var parameters []byte
	var err error
	if request != nil && request.Method == http.MethodGet {
		parameters, err = publicGETParameters(request)
	} else {
		parameters, err = publicPOSTParameters(writer, request)
	}
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	defer clear(parameters)
	handler.serve(writer, request, parameters)
}

func (handler *Handler) serve(writer http.ResponseWriter, request *http.Request, parameters []byte) {
	publicID, slug, kind, identityOK := publicIdentity(request)
	if !identityOK || !publicHeadersValid(request) {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	token, err := publicCapability(request)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	defer clear(token)
	clientIP, err := handler.runtime.ClientAddress(request)
	if err != nil || clientIP == "" {
		handler.runtime.Failure(writer, request, httptransport.ErrUnavailable)
		return
	}
	result, err := handler.public.Execute(request.Context(), publicationapp.PublicRequest{
		PublicID: publicID, ExpectedKind: kind, Slug: slug, TrustedClientIP: clientIP,
		CapabilityToken: token, Parameters: parameters})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	if err = handler.runtime.PublicJSON(writer, request, result.Body); err != nil {
		handler.runtime.Report(errors.New("analytics public response write failed"))
	}
}
