package compatibilityhttp

import (
	"errors"
	"net/http"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
)

func (handler *Handler) Authorize(writer http.ResponseWriter, request *http.Request) {
	key, err := legacyKey(request)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	mapping, err := handler.application.AuthorizeLegacy(request.Context(), key)
	if err != nil || !exactAdmission(mapping, key) {
		if err == nil {
			err = compatibilityapp.ErrUnavailable
		}
		handler.runtime.Failure(writer, request, err)
		return
	}
	parameters, err := legacyParameters(writer, request)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	defer clear(parameters)
	clientIP, err := handler.runtime.ClientAddress(request)
	if err != nil || clientIP == "" {
		handler.runtime.Failure(writer, request, httptransport.ErrUnavailable)
		return
	}
	result, err := handler.public.Execute(request.Context(), publicationapp.PublicRequest{
		PublicID: mapping.PublicID, ExpectedKind: publicationapp.ResourceQuery,
		Slug: "legacy", TrustedClientIP: clientIP, Parameters: parameters})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	if err = handler.runtime.PublicJSON(writer, request, result.Body); err != nil {
		handler.runtime.Report(errors.New("analytics legacy response write failed"))
	}
}

func exactAdmission(mapping compatibilityapp.LegacyMapping, key compatibilityapp.LegacyKey) bool {
	return mapping.Key == key && mapping.State == "active" && mapping.PublicID != [16]byte{} &&
		mapping.HitCount > 0 && mapping.LastHitAt != nil
}

func legacyParameters(writer http.ResponseWriter, request *http.Request) ([]byte, error) {
	if request.Method == http.MethodGet {
		return getParameters(request)
	}
	return postParameters(writer, request)
}
