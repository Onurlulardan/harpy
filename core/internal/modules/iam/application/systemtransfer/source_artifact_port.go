package systemtransfer

import (
	"context"
	"io"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

type ExactImportSource struct {
	TenantID, ActorID, SourceID uuid.UUID
	Resource                    publicresource.Resource
	Format                      string
	Artifact                    SourceArtifact
}

func (source ExactImportSource) Validate() error {
	if source.TenantID == uuid.Nil || source.ActorID == uuid.Nil ||
		source.SourceID == uuid.Nil || source.SourceID.Version() != 7 ||
		!source.Resource.Valid() ||
		(source.Format != "csv" && source.Format != "jsonl") ||
		source.Artifact.Validate() != nil {
		return ErrInvalidAttempt
	}
	return nil
}

// ExactImportSourceOpener returns authenticated plaintext for one exact object
// version. Implementations verify FCART header/AAD, ciphertext and plaintext
// sizes and SHA-256 digests before returning EOF; tamper evidence fails closed.
type ExactImportSourceOpener interface {
	OpenExactImportSource(context.Context, ExactImportSource) (io.ReadCloser, error)
}
