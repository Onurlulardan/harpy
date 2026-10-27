package licenseauthoritysource

import (
	"context"

	licenseauthority "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
)

type SnapshotFile struct{ path string }

func NewSnapshotFile(path string) (*SnapshotFile, error) {
	if !validArtifactPath(path) {
		return nil, ErrInvalidConfiguration
	}
	return &SnapshotFile{path: path}, nil
}

func (source *SnapshotFile) IssueSnapshot(
	ctx context.Context, _ licenseauthority.IssueRequest,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return readStableArtifact(source.path, licenseauthority.MaximumDocumentBytes, true)
}

var _ licenseauthority.Source = (*SnapshotFile)(nil)
