package iam

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type postRestoreChallengeData struct {
	OperationID string                       `json:"operation_id"`
	PublicKey   postrestore.PublicKeyRequest `json:"public_key"`
	ExpiresAt   time.Time                    `json:"expires_at"`
}

type postRestoreChallengeEnvelope struct {
	Success bool                     `json:"success"`
	Data    postRestoreChallengeData `json:"data"`
}

type postRestoreReactivateData struct {
	UserID       string    `json:"user_id"`
	TenantID     string    `json:"tenant_id"`
	SessionState string    `json:"session_state"`
	Token        string    `json:"token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type postRestoreReactivateEnvelope struct {
	Success bool                      `json:"success"`
	Data    postRestoreReactivateData `json:"data"`
}

func (handler *PostRestoreHTTP) writePostRestoreJSON(
	w http.ResponseWriter, r *http.Request, status int, value any,
) {
	w.Header().Set("Cache-Control", "no-store")
	if err := httpwire.WriteJSON(w, status, authHTTPCorrelationID(r), value, false); err != nil {
		handler.report(fmt.Errorf("write IAM post-restore response: %w", err))
	}
}
