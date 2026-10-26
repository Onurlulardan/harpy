package platformops

import (
	"net/http"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

type restoreRequestDocument struct {
	TargetDisposition       platformopsapi.TargetDisposition `json:"target_disposition"`
	SafetyRecoverySetID     *string                          `json:"safety_recovery_set_id"`
	SafetyRecoverySetSHA256 *string                          `json:"safety_recovery_set_sha256"`
	WaiverReason            *string                          `json:"waiver_reason"`
	ProbeSHA256             *string                          `json:"probe_sha256"`
	DestructiveConfirmed    *bool                            `json:"destructive_confirmed"`
}

func decodeRestoreRequest(
	w http.ResponseWriter,
	r *http.Request,
) (restoreRequestDocument, error) {
	var document restoreRequestDocument
	if err := httpinput.DecodeJSON(w, r, 4096, &document); err != nil {
		return restoreRequestDocument{}, err
	}
	if document.DestructiveConfirmed == nil ||
		(document.TargetDisposition != platformopsapi.TargetExistingReadable &&
			document.TargetDisposition != platformopsapi.TargetEmpty &&
			document.TargetDisposition != platformopsapi.TargetUnreadable) {
		return restoreRequestDocument{}, httpinput.ErrInvalid
	}
	return document, nil
}

func nullableString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
