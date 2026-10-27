package dashboardexportartifact

import (
	"context"
	"errors"
	"fmt"
	"io"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

func (runtime *Runtime) PrepareReserved(
	ctx context.Context, pin exportapp.Pin, generation int64,
	leaseOwner uuid.UUID, spool exportapp.EncryptedSpool, renderer exportapp.Renderer,
) (exportapp.PendingArtifact, error) {
	identity := exportapp.SpoolReservationIdentity{ExportID: pin.ExportID,
		AttemptGeneration: generation, LeaseOwnerID: leaseOwner}
	if runtime == nil || ctx == nil || ctx.Err() != nil || renderer == nil || spool == nil ||
		spool.Identity() != identity || pin.ExportID.Version() != 7 ||
		pin.TenantID == uuid.Nil || leaseOwner == uuid.Nil || generation < 1 ||
		generation > exportapp.MaximumAttempts || pin.MaximumPlaintextBytes < 1 ||
		pin.MaximumPlaintextBytes > exportapp.MaximumPlaintextBytes {
		return nil, exportapp.ErrInvalid
	}
	keyIdentity := fcartv1.KeyIdentity{TenantID: pin.TenantID,
		ArtifactID: pin.ExportID, AttemptGeneration: uint64(generation)}
	key, envelope, err := runtime.keys.Generate(ctx, keyIdentity)
	if err != nil {
		return nil, errors.Join(exportapp.ErrUnavailable, err)
	}
	defer clear(key[:])
	format, err := artifactFormat(pin.Format)
	if err != nil {
		return nil, err
	}
	header, err := fcartv1.NewHeader(pin.ExportID, uint64(generation), format)
	if err != nil {
		return nil, errors.Join(exportapp.ErrUnavailable, err)
	}
	destination, err := spool.OpenWriter()
	if err != nil {
		return nil, errors.Join(exportapp.ErrUnavailable, err)
	}
	plaintext, renderDestination := io.Pipe()
	evidence := newPlaintextEvidenceWriter(renderDestination, pin.MaximumPlaintextBytes)
	type renderResult struct {
		stats exportapp.RenderStats
		err   error
	}
	rendered := make(chan renderResult, 1)
	go func() {
		stats, renderErr := renderer.Render(ctx, pin, evidence)
		_ = renderDestination.CloseWithError(renderErr)
		rendered <- renderResult{stats: stats, err: renderErr}
	}()
	encrypted, encryptErr := fcartv1.Encrypt(destination, plaintext, header, key,
		fcartv1.NewCryptographicRandomSource())
	if encryptErr != nil {
		_ = plaintext.CloseWithError(encryptErr)
	} else {
		_ = plaintext.Close()
	}
	result := <-rendered
	closeErr := destination.Close()
	plaintextBytes, plainDigest := evidence.evidence()
	if result.err != nil || encryptErr != nil || closeErr != nil ||
		plaintextBytes < 1 || int64(encrypted.PlaintextBytes) != plaintextBytes ||
		result.stats.Rows < 0 || result.stats.Rows > pin.MaximumRows {
		return nil, errors.Join(exportapp.ErrUnavailable, result.err,
			encryptErr, closeErr)
	}
	draft := exportapp.ArtifactDraft{AttemptGeneration: generation,
		ObjectKey: fmt.Sprintf("%s/v1/%s/%d.fcart", runtime.prefix,
			pin.ExportID.String(), generation), Envelope: envelope.OwnedCopy(),
		HeaderSHA256:     encrypted.HeaderSHA256,
		CiphertextSHA256: encrypted.CiphertextSHA256, PlaintextSHA256: plainDigest,
		CiphertextBytes: int64(encrypted.CiphertextBytes),
		PlaintextBytes:  plaintextBytes, Rows: result.stats.Rows,
		ContentType: pin.Format.ContentType(), Filename: fmt.Sprintf(
			"dashboard-%s-%s.%s", pin.DashboardID, pin.ExportID, pin.Format.Extension())}
	return &pending{runtime: runtime, lease: spool, draft: draft}, nil
}
