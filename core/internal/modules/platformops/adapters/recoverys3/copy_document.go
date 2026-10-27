package recoverys3

import (
	"context"
	"io"
)

func CopyDocumentVersion(
	ctx context.Context,
	source *Reader,
	target *Writer,
	sourceIdentity ObjectIdentity,
	targetKey string,
) (WriteEvidence, error) {
	if source == nil || target == nil || ctx == nil || ctx.Err() != nil ||
		!validObjectKey(targetKey) {
		return WriteEvidence{}, ErrObject
	}
	if err := source.DownloadVerified(ctx, sourceIdentity, io.Discard); err != nil {
		return WriteEvidence{}, err
	}
	body, err := source.OpenExact(
		ctx, sourceIdentity.Key, sourceIdentity.VersionID,
	)
	if err != nil {
		return WriteEvidence{}, err
	}
	defer body.Close()
	evidence, writeErr := target.PutVerified(ctx, WriteRequest{
		Key: targetKey, Body: body, ContentLength: sourceIdentity.Bytes,
		ExpectedSHA256: sourceIdentity.SHA256,
		ExpectedBytes:  sourceIdentity.Bytes,
	})
	if writeErr != nil {
		return WriteEvidence{}, ErrObject
	}
	return evidence, nil
}
