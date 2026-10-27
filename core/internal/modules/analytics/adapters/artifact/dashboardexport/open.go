package dashboardexportartifact

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func (runtime *Runtime) Open(
	ctx context.Context, stored exportapp.StoredContent,
) (io.ReadCloser, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil || stored.ObjectVersion == "" ||
		stored.Draft.ObjectKey == "" || stored.Draft.PlaintextBytes < 1 ||
		stored.Draft.CiphertextBytes < 1 || stored.Draft.Envelope.Validate() != nil {
		return nil, exportapp.ErrInvalid
	}
	format, err := artifactFormat(stored.Format)
	if err != nil {
		return nil, err
	}
	identity := fcartv1.KeyIdentity{TenantID: stored.TenantID,
		ArtifactID:        stored.ExportID,
		AttemptGeneration: uint64(stored.Draft.AttemptGeneration)}
	key, err := runtime.keys.Unwrap(ctx, identity, stored.Draft.Envelope)
	if err != nil {
		return nil, errors.Join(exportapp.ErrUnavailable, err)
	}
	header, err := fcartv1.NewHeader(stored.ExportID,
		uint64(stored.Draft.AttemptGeneration), format)
	if err != nil {
		clear(key[:])
		return nil, errors.Join(exportapp.ErrUnavailable, err)
	}
	object, err := runtime.objects.OpenExact(ctx, artifactobject.Expected{
		Identity: artifactobject.Identity{Key: stored.Draft.ObjectKey,
			Version: stored.ObjectVersion}, CiphertextBytes: stored.Draft.CiphertextBytes,
		CiphertextSHA256: stored.Draft.CiphertextSHA256,
		HeaderSHA256:     stored.Draft.HeaderSHA256})
	if err != nil {
		clear(key[:])
		return nil, errors.Join(exportapp.ErrUnavailable, err)
	}
	plaintext, destination := io.Pipe()
	go streamDecryptedArtifact(destination, object, header, key, stored.Draft)
	return plaintext, nil
}

func streamDecryptedArtifact(destination *io.PipeWriter, object io.ReadCloser,
	header fcartv1.Header, key fcartv1.DEK, draft exportapp.ArtifactDraft,
) {
	digestWriter := sha256.New()
	result, decryptErr := fcartv1.Decrypt(io.MultiWriter(destination, digestWriter), object,
		fcartv1.CiphertextMetadata{Header: header,
			HeaderSHA256: draft.HeaderSHA256, CiphertextBytes: uint64(draft.CiphertextBytes),
			CiphertextSHA256: draft.CiphertextSHA256}, key)
	closeErr := object.Close()
	clear(key[:])
	var digest [sha256.Size]byte
	copy(digest[:], digestWriter.Sum(nil))
	if decryptErr != nil || closeErr != nil ||
		int64(result.PlaintextBytes) != draft.PlaintextBytes ||
		digest != draft.PlaintextSHA256 {
		_ = destination.CloseWithError(errors.Join(
			exportapp.ErrUnavailable, decryptErr, closeErr))
		return
	}
	_ = destination.Close()
}
