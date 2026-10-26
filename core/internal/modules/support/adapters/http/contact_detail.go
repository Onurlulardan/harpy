package supporthttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (handler *Handler) getContact(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopePlatform,
		application.PermissionPlatformContactRead, application.OperationContactRead)
	if err == nil && len(request.URL.Query()) != 0 {
		err = application.ErrInvalid
	}
	id, idErr := pathUUID(request)
	if err == nil {
		err = idErr
	}
	if err == nil {
		var message domain.ContactMessage
		message, err = handler.applications.Contact.GetContact(request.Context(), access, domain.ID(id))
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, contactView(message))
			return
		}
	}
	handler.writeFailure(response, request, err)
}

func (handler *Handler) updateContact(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopePlatform,
		application.PermissionPlatformContactManage, application.OperationContactUpdate)
	id, idErr := pathUUID(request)
	if err == nil {
		err = idErr
	}
	if err == nil && len(request.URL.Query()) != 0 {
		err = application.ErrInvalid
	}
	var document contactUpdateDocument
	var canonical []byte
	if err == nil {
		canonical, err = decodeObject(response, request, &document)
	}
	var key string
	var digest [32]byte
	if err == nil {
		key, digest, err = commandEvidence(request, canonical)
	}
	if err == nil {
		var message domain.ContactMessage
		message, err = handler.applications.Contact.UpdateContact(request.Context(), access,
			domain.ContactUpdate{ID: domain.ID(id), Status: document.Status, Notes: document.Notes,
				ExpectedRevision: document.ExpectedRevision, ActorID: domain.ID(access.ActorID()),
				IdempotencyKey: key, RequestSHA256: digest})
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, contactView(message))
			return
		}
	}
	handler.writeFailure(response, request, err)
}
