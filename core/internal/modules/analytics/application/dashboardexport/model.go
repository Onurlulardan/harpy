package dashboardexport

import (
	"crypto/sha256"
	"time"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	"github.com/google/uuid"
)

const (
	MaximumWidgets              = 30
	MaximumRows           int64 = 100_000
	MaximumPlaintextBytes       = 64 << 20
	MaximumAttempts       int64 = 3
	ExportLifetime              = 24 * time.Hour
)

type Snapshot struct {
	ExportID, DashboardID, DashboardVersionID, JobID uuid.UUID
	State, StoredState                               State
	Format                                           Format
	Filename, ContentType, FailureCode               string
	RowCount, PlaintextBytes                         int64
	CancellationRequested                            bool
	CreatedAt, UpdatedAt, ExpiresAt                  time.Time
}

type PinnedWidget struct {
	WidgetID, QueryVersionID uuid.UUID
	Position                 uint16
	Parameters, Output       []byte
}

type Pin struct {
	TenantID, ActorID, ExportID, DashboardID, DashboardVersionID uuid.UUID
	TokenVersion, AuthorizationVersion, AuthorizationClock       int64
	DashboardManifestSHA256, PinSHA256                           [sha256.Size]byte
	Format                                                       Format
	Parameters                                                   []byte
	Widgets                                                      []PinnedWidget
	MaximumRows, MaximumPlaintextBytes                           int64
	ExpiresAt                                                    time.Time
}

type ArtifactDraft struct {
	AttemptGeneration                               int64
	ObjectKey                                       string
	Envelope                                        fcartv1.KeyEnvelope
	HeaderSHA256, CiphertextSHA256, PlaintextSHA256 [sha256.Size]byte
	CiphertextBytes, PlaintextBytes, Rows           int64
	ContentType, Filename                           string
}

type UploadReceipt struct {
	ObjectVersion, ProviderRequestID string
	UploadedAt                       time.Time
}

type StoredContent struct {
	Pin
	Draft         ArtifactDraft
	ObjectVersion string
	ReadyAt       time.Time
}

type Content struct {
	Reader      ReadCloser
	Length      int64
	ContentType string
	Filename    string
}
