package lifecyclehttp

import (
	"net/http"

	"github.com/google/uuid"
)

type purgeDocument struct {
	OperationID       uuid.UUID `json:"operation_id"`
	BackupReceiptID   uuid.UUID `json:"backup_receipt_id"`
	ApprovalReceiptID uuid.UUID `json:"approval_receipt_id"`
	Reason            string    `json:"reason"`
	ConfirmationSlug  string    `json:"confirmation_slug"`
}

func (handler *Handler) PurgeRoute() Route {
	return route(http.MethodPost, "/api/catalog/installations/{installation_id}/purges", handler.purge)
}

func (handler *Handler) purge(writer http.ResponseWriter, request *http.Request) {
	installationID, err := pathV7(request, "installation_id")
	var document purgeDocument
	if err == nil {
		err = requireNoQuery(request)
	}
	if err == nil {
		err = decodeBody(request, &document)
	}
	key := ""
	if err == nil {
		key, err = idempotencyKey(request)
	}
	if err != nil || !v7(document.OperationID) || !v7(document.BackupReceiptID) ||
		!v7(document.ApprovalReceiptID) {
		handler.failure(writer, request, ErrInvalid)
		return
	}
	purpose := installationID.String() + ":" + document.OperationID.String() +
		":" + document.ApprovalReceiptID.String()
	access, err := handler.authorize(request, PermissionPurge, "catalog.installation.purge", purpose)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.Purge(request.Context(), access, PurgeInput{
		InstallationID: installationID, OperationID: document.OperationID,
		BackupReceiptID:   document.BackupReceiptID,
		ApprovalReceiptID: document.ApprovalReceiptID,
		StepUpAt:          access.StepUpAt(), Reason: document.Reason,
		ConfirmationSlug: document.ConfirmationSlug, IdempotencyKey: key,
	})
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusAccepted, value)
}

func v7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 && value.Variant() == uuid.RFC4122
}
