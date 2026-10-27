package publicationhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	publicationdomain "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
)

func (handler *Handler) Create(writer http.ResponseWriter, request *http.Request) {
	key, keyOK := httptransport.IdempotencyKey(request)
	if !keyOK || request.URL.RawQuery != "" {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	document, err := httptransport.Decode[createDocument](writer, request)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	permission, permissionOK := createPermission(document.Audience)
	if !permissionOK {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	access, err := handler.runtime.AccessFor(request, permission, httptransport.OperationPublicationCreate)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	versionID, versionOK := httptransport.WireV7(document.ResourceVersionID)
	if !versionOK {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	result, err := handler.application.CreatePublication(request.Context(), access, publicationapp.CreateRequest{
		ResourceKind: publicationapp.ResourceKind(document.ResourceKind), ResourceVersionID: versionID,
		Audience: publicationdomain.Audience(document.Audience), CachePolicy: publicationdomain.CachePolicy(document.CachePolicy),
		IdempotencyKey: key})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	writer.Header().Set("Location", publicationsPath+"/"+result.ReleaseID.String())
	handler.runtime.JSON(writer, request, http.StatusCreated, present(result))
}

func createPermission(audience string) (httptransport.Permission, bool) {
	switch publicationdomain.Audience(audience) {
	case publicationdomain.AudienceTenant:
		return httptransport.PermissionPublicationRead, true
	case publicationdomain.AudiencePublic:
		return httptransport.PermissionPublicationPublishPublic, true
	case publicationdomain.AudienceCapability:
		return httptransport.PermissionPublicationPublishCapability, true
	default:
		return "", false
	}
}
