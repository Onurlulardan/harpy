package billing

import (
	"fmt"
	"net/http"
	"time"

	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type platformLicenseView struct {
	LicenseID        string                              `json:"license_id"`
	EnvironmentID    string                              `json:"environment_id"`
	SnapshotSequence int64                               `json:"snapshot_sequence"`
	SnapshotSHA256   string                              `json:"snapshot_sha256"`
	State            string                              `json:"state"`
	ReleaseChannel   string                              `json:"release_channel"`
	Features         map[string]bool                     `json:"features"`
	Limits           map[string]platformlicenseapp.Limit `json:"limits"`
	DatabaseNow      time.Time                           `json:"database_now"`
}

type platformLicenseRefreshAcceptedView struct {
	OperationID string `json:"operation_id"`
	State       string `json:"state"`
}

type platformLicenseRefreshOperationView struct {
	OperationID      string     `json:"operation_id"`
	State            string     `json:"state"`
	SnapshotSequence *int64     `json:"snapshot_sequence"`
	SnapshotSHA256   *string    `json:"snapshot_sha256"`
	FailureCode      *string    `json:"failure_code"`
	CreatedAt        time.Time  `json:"created_at"`
	CompletedAt      *time.Time `json:"completed_at"`
}

func (handler *PlatformLicenseHTTPHandler) writeSuccess(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	data any,
) {
	if err := httpwire.WriteJSON(
		w, status, correlationID(r), successEnvelope{Success: true, Data: data}, false,
	); err != nil {
		handler.report(fmt.Errorf("write platform license response: %w", err))
	}
}
