package platformops

import (
	"fmt"
	"net/http"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type platformOpsSuccessEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type platformOpsPageMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

type platformOpsPageEnvelope struct {
	Success bool                `json:"success"`
	Data    any                 `json:"data"`
	Meta    platformOpsPageMeta `json:"meta"`
}

type recoveryOperationView struct {
	OperationID    string     `json:"operation_id"`
	Kind           string     `json:"kind"`
	State          string     `json:"state"`
	KitClaimState  string     `json:"kit_claim_state"`
	ArtifactSHA256 *string    `json:"artifact_sha256"`
	ExpiresAt      *time.Time `json:"expires_at"`
	FailureCode    *string    `json:"failure_code"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (handler *HTTPHandler) writePage(
	w http.ResponseWriter, r *http.Request, data any, meta platformOpsPageMeta,
) {
	err := httpwire.WriteJSON(w, http.StatusOK, platformOpsCorrelationID(r),
		platformOpsPageEnvelope{Success: true, Data: data, Meta: meta}, false)
	if err != nil {
		handler.report(fmt.Errorf("write platform operations page: %w", err))
	}
}

func representRecoveryOperation(value platformopsapi.RecoveryControlOperation) recoveryOperationView {
	return recoveryOperationView{
		OperationID: value.OperationID, Kind: value.Kind, State: value.State,
		KitClaimState: value.KitClaimState, ArtifactSHA256: optionalString(value.ArtifactSHA256),
		ExpiresAt: value.ExpiresAt, FailureCode: optionalString(value.FailureCode),
		CreatedAt: value.CreatedAt,
	}
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (handler *HTTPHandler) writeSuccess(
	w http.ResponseWriter, r *http.Request, status int, data any,
) {
	err := httpwire.WriteJSON(w, status, platformOpsCorrelationID(r),
		platformOpsSuccessEnvelope{Success: true, Data: data}, false)
	if err != nil {
		handler.report(fmt.Errorf("write platform operations HTTP response: %w", err))
	}
}

func platformOpsCorrelationID(r *http.Request) requestmeta.CorrelationID {
	if r != nil {
		if value, found := requestmeta.CorrelationIDFromContext(r.Context()); found {
			return value
		}
	}
	return requestmeta.EntropyFailureCorrelationID
}
