package systemtransfer

import (
	"context"
	"crypto/sha256"
	"io"
	"time"

	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
)

type exportSnapshotsFixture struct {
	snapshot ExportSnapshot
	active   bool
	closed   bool
}

type exportRowsFixture struct {
	rows  [][]tabulartransfer.Cell
	index int
	eof   bool
}

func (rows *exportRowsFixture) Next() ([]tabulartransfer.Cell, error) {
	if rows.index == len(rows.rows) {
		rows.eof = true
		return nil, io.EOF
	}
	value := rows.rows[rows.index]
	rows.index++
	return value, nil
}

func (snapshots *exportSnapshotsFixture) WithPinnedSnapshot(
	ctx context.Context,
	_ ExportAttempt,
	read func(context.Context, ExportSnapshot) error,
) error {
	snapshots.active = true
	err := read(ctx, snapshots.snapshot)
	snapshots.active = false
	snapshots.closed = true
	return err
}

type exportSealerFixture struct {
	snapshots  *exportSnapshotsFixture
	spool      *exportSpoolFixture
	rows       *exportRowsFixture
	store      *exportStoreFixture
	now        time.Time
	called     int
	plaintext  []byte
	afterBegin error
}

func (sealer *exportSealerFixture) SealExport(
	ctx context.Context,
	request ExportArtifactSealRequest,
	begin ExportArtifactBegin,
	commit ExportArtifactCommit,
) error {
	sealer.called++
	if request.Validate() != nil || sealer.snapshots.active ||
		!sealer.snapshots.closed || !sealer.spool.writerClosed || !sealer.rows.eof {
		return ErrExportStateConflict
	}
	plaintext, err := io.ReadAll(request.Plaintext)
	if err != nil || len(plaintext) == 0 {
		return ErrExportStateConflict
	}
	sealer.plaintext = plaintext
	objectKey := "iam/exports/artifact.fcart"
	requestSHA256 := sha256.Sum256(plaintext)
	draft := ExportArtifactDraftEvidence{KeyID: "iam-export-key-1",
		HeaderSHA256: [32]byte{1}, CiphertextSHA256: [32]byte{2},
		CiphertextBytes: int64(len(plaintext)) + 64,
		WrappedDEK:      make([]byte, 32), WrapNonce: [12]byte{1}}
	if err = begin(ctx, ExportArtifactBeginEvidence{EffectID: request.EffectID,
		ObjectKey: objectKey, RequestSHA256: requestSHA256,
		StartedAt: sealer.now, Draft: draft}); err != nil {
		return err
	}
	if sealer.store == nil || len(sealer.store.starts) != 1 {
		return ErrExportStateConflict
	}
	if sealer.afterBegin != nil {
		return sealer.afterBegin
	}
	metadata := ExportArtifactMetadata{ObjectKey: "iam/exports/artifact.fcart",
		ObjectVersion: "version-1", KeyID: "iam-export-key-1",
		ProviderRequestID: "provider-request-1", EffectID: request.EffectID,
		EffectReceiptID: request.EffectReceiptID,
		HeaderSHA256:    draft.HeaderSHA256, CiphertextSHA256: draft.CiphertextSHA256,
		RequestSHA256: requestSHA256, ResultSHA256: sha256.Sum256([]byte("stored")),
		CiphertextBytes: draft.CiphertextBytes, WrappedDEK: draft.WrappedDEK,
		WrapNonce: draft.WrapNonce, AttemptGeneration: request.AttemptGeneration,
		SealedAt: sealer.now}
	return commit(ctx, metadata)
}

func fixtureExportSnapshot(
	attempt ExportAttempt,
	now time.Time,
	rows *exportRowsFixture,
) ExportSnapshot {
	return ExportSnapshot{CatalogVersionID: attempt.CatalogVersionID,
		CatalogSHA256: attempt.CatalogSHA256, QuerySHA256: attempt.QuerySHA256,
		StartedAt: now, Columns: []string{"id", "email"}, Rows: rows}
}

func fixtureExportRows() *exportRowsFixture {
	return &exportRowsFixture{rows: [][]tabulartransfer.Cell{{
		{Kind: tabulartransfer.CellText, Text: fixtureRecordID.String()},
		{Kind: tabulartransfer.CellText, Text: "person@example.com"},
	}}}
}
