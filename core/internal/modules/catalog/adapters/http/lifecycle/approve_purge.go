package lifecyclehttp

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/google/uuid"
)

type purgeApprovalDocument struct {
	OperationID         uuid.UUID `json:"operation_id"`
	BackupReceiptID     uuid.UUID `json:"backup_receipt_id"`
	RequestedBy         uuid.UUID `json:"requested_by"`
	Reason              string    `json:"reason"`
	ConfirmationSlug    string    `json:"confirmation_slug"`
	PurgeIdempotencyKey string    `json:"purge_idempotency_key"`
}

func (handler *Handler) ApprovePurgeRoute() Route {
	return route(
		http.MethodPost,
		"/api/catalog/installations/{installation_id}/purge-approvals",
		handler.approvePurge,
	)
}

func (handler *Handler) approvePurge(writer http.ResponseWriter, request *http.Request) {
	installationID, err := pathV7(request, "installation_id")
	var document purgeApprovalDocument
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
		!v7(document.RequestedBy) || document.Reason == "" ||
		document.ConfirmationSlug == "" || document.PurgeIdempotencyKey == "" {
		handler.failure(writer, request, ErrInvalid)
		return
	}
	purpose := approvalPurpose(installationID, document)
	access, err := handler.authorize(
		request, PermissionPurge, "catalog.installation.purge.approve", purpose,
	)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.ApprovePurge(
		request.Context(), access, cataloghttp.PurgeApprovalInput{
			InstallationID: installationID, OperationID: document.OperationID,
			BackupReceiptID: document.BackupReceiptID, RequestedBy: document.RequestedBy,
			StepUpAt: access.StepUpAt(), Reason: document.Reason,
			ConfirmationSlug:       document.ConfirmationSlug,
			PurgeIdempotencyKey:    document.PurgeIdempotencyKey,
			ApprovalIdempotencyKey: key,
		},
	)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusCreated, value)
}

func approvalPurpose(
	installationID uuid.UUID,
	document purgeApprovalDocument,
) string {
	hash := sha256.New()
	for _, value := range []string{
		installationID.String(), document.OperationID.String(),
		document.BackupReceiptID.String(), document.RequestedBy.String(),
		document.Reason, document.ConfirmationSlug, document.PurgeIdempotencyKey,
	} {
		_, _ = hash.Write([]byte{byte(len(value) >> 8), byte(len(value))})
		_, _ = hash.Write([]byte(value))
	}
	return installationID.String() + ":" + hex.EncodeToString(hash.Sum(nil))
}
