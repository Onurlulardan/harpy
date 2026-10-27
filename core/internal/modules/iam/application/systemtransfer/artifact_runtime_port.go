package systemtransfer

import (
	"context"
	"io"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

type SourceRegistration func(
	context.Context,
	SourceArtifact,
) (SourceRecord, error)

type SourceSealRequest struct {
	TenantID, ActorID uuid.UUID
	SourceID          uuid.UUID
	Resource          publicresource.Resource
	Format            string
	IdempotencySHA256 [32]byte
	UploadedAt        time.Time
	Body              io.Reader
	MaximumBytes      int64
}

type ExportOpenRequest struct {
	TenantID uuid.UUID
	Resource publicresource.Resource
	ExportID uuid.UUID
	Artifact ExportArtifact
}

type ArtifactRuntime interface {
	SealImportSource(context.Context, SourceSealRequest,
		SourceRegistration) (SourceRecord, error)
	OpenExport(context.Context, ExportOpenRequest) (io.ReadCloser, error)
}
